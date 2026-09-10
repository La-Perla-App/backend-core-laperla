package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"connectrpc.com/connect"
	"github.com/La-Perla-App/backend-core-laperla/pkg/api/auth"
	"github.com/google/uuid"
)

func NewRequest[T any](generalParams GeneralParams, request *T) (*connect.Request[T], error) {
	rpcRequest := connect.NewRequest(request)
	if err := SetGeneralParamsHeader(generalParams, rpcRequest.Header()); err != nil {
		return rpcRequest, err
	}
	if token := strings.TrimSpace(generalParams.SessionToken); token != "" && strings.TrimSpace(rpcRequest.Header().Get("Authorization")) == "" {
		rpcRequest.Header().Set("Authorization", "Bearer "+token)
	}
	return rpcRequest, nil
}

func GeneralParamsFromConnectRequest[T any](request *connect.Request[T]) (GeneralParams, error) {
	return GeneralParamsFromHeaders(request.Header())
}

func GeneralParamsFromHeaders(headers http.Header) (GeneralParams, error) {
	var generalParams GeneralParams
	header := headers.Get(generalParamsHeaderKey)
	if header != "" {
		jsonGeneralParams, err := connect.DecodeBinaryHeader(header)
		if err != nil {
			return generalParams, err
		}
		if err := json.Unmarshal(jsonGeneralParams, &generalParams); err != nil {
			return generalParams, err
		}
	} else {
		generalParams = generalParamsFromHTTPRequest(headers)
	}
	return generalParams, nil
}

func SetGeneralParamsHeader(generalParams GeneralParams, headers http.Header) error {
	jsonGeneralParams, err := json.Marshal(generalParams)
	if err != nil {
		return err
	}
	headers.Set(generalParamsHeaderKey, connect.EncodeBinaryHeader(jsonGeneralParams))
	return nil
}

func SetGeneralParams[T any](generalParams GeneralParams, request *connect.Request[T]) error {
	return SetGeneralParamsHeader(generalParams, request.Header())
}

func SetResponseInfoHeader(info ResponseInfo, headers http.Header) error {
	jsonInfo, err := json.Marshal(info)
	if err != nil {
		return err
	}
	headers.Set(ResponseInfoHeaderKey, connect.EncodeBinaryHeader(jsonInfo))
	return nil
}

func UpdateResponseInfoMessage(message string, headers http.Header) error {
	return UpdateResponseInfo(headers, func(responseInfo ResponseInfo) ResponseInfo {
		responseInfo.Message = message
		return responseInfo
	})
}

func UpdateResponseInfoErrorMessage(err error, headers http.Header) error {
	msg := err.Error()
	if err, ok := err.(*connect.Error); ok {
		msg = err.Message()
	}
	UpdateResponseInfoMessage(msg, headers)
	return err
}

func UpdateResponseInfoErrorMessageFromCode(code APIErrorCode, headers http.Header) error {
	msg := GetErrorMessageFromCode(code)
	UpdateResponseInfoMessage(msg, headers)

	connectCode := connect.Code(code)
	if connectCode > connect.CodeResourceExhausted {
		connectCode = connect.CodeInternal
	}

	_ = UpdateResponseInfo(headers, func(responseInfo ResponseInfo) ResponseInfo {
		responseInfo.APIErrorCode = code
		return responseInfo
	})

	return connect.NewError(connectCode, errors.New(msg))
}

func UpdateResponseInfo(headers http.Header, updateFn func(responseInfo ResponseInfo) ResponseInfo) error {
	responseInfo, _ := ResponseInfoFromHeaders(headers)
	responseInfo = updateFn(responseInfo)
	return SetResponseInfoHeader(responseInfo, headers)
}

func UpdateResponseInfoSessionToken(sessionId string, headers http.Header) error {
	return UpdateResponseInfo(headers, func(responseInfo ResponseInfo) ResponseInfo {
		responseInfo.SessionToken = sessionId
		return responseInfo
	})
}

func UpdateResponseInfoType(infoType ResponseType, headers http.Header) error {
	return UpdateResponseInfo(headers, func(responseInfo ResponseInfo) ResponseInfo {
		responseInfo.Type = infoType
		return responseInfo
	})
}

func ResponseInfoFromRequest[T any](request *connect.Request[T]) (ResponseInfo, error) {
	return ResponseInfoFromHeaders(request.Header())
}

func ResponseInfoFromHeaders(headers http.Header) (ResponseInfo, error) {
	var info ResponseInfo
	header := headers.Get(ResponseInfoHeaderKey)
	jsonResponseInfo, err := connect.DecodeBinaryHeader(header)
	if err != nil {
		return info, err
	}
	if err := json.Unmarshal(jsonResponseInfo, &info); err != nil {
		return info, err
	}
	return info, nil
}

func GeneralParamsFromHTTPRequest(r *http.Request) GeneralParams {
	return generalParamsFromHTTPRequest(r.Header)
}

func generalParamsFromHTTPRequest(headers http.Header) GeneralParams {
	if headers.Get(generalParamsHeaderKey) != "" {
		generalParams, err := GeneralParamsFromHeaders(headers)
		if err == nil {
			return generalParams
		}
	}

	params := GeneralParams{
		IANATimezone: "America/Caracas",
	}

	params.Lang = SupportedLanguages[0]
	if acceptLanguage := headers.Get("Accept-Language"); acceptLanguage != "" {
		requestLanguages := ParseAcceptLanguage(acceptLanguage)
		for _, lang := range requestLanguages {
			if slices.Contains(SupportedLanguages, lang.ShortTag) {
				params.Lang = lang.ShortTag
				break
			}
		}
	}

	if iana := headers.Get("X-Timezone"); iana != "" {
		params.IANATimezone = iana
	}

	params.Platform = headers.Get("X-Platform")
	if strings.TrimSpace(params.Platform) == "" {
		params.Platform = checkPlatformByUserAgent(headers.Get("User-Agent"))
	}
	params.ClientVersion = headers.Get("X-Client-Version")
	if strings.TrimSpace(params.ClientVersion) == "" {
		params.ClientVersion = "1.0.0"
	}
	switch params.Platform {
	case PlatformWeb:
	case PlatformAndroid:
	case PlatformIOS:
	default:
		params.Platform = PlatformUnknown
	}

	if cookieHeader := headers.Get("Cookie"); cookieHeader != "" {
		cookies, _ := http.ParseCookie(cookieHeader)
		for _, cookie := range cookies {
			if cookie.Name == ClientIDCookieName {
				params.ClientId = cookie.Value
			}
		}
	}

	if id := headers.Get("X-Client-Id"); id != "" {
		params.ClientId = id
	}

	if params.ClientId == "" {
		params.ClientId = uuid.NewString()
	}

	token, err := auth.GetTokenFromHeader(headers)
	if err != nil {
		log.Println("AUTH WARN: authorization token not found")
	}
	params.SessionToken = token

	return params
}

func CheckSessionFromConnectRequest[T any](request *connect.Request[T]) (*auth.SessionData, error) {
	generalParams, err := GeneralParamsFromConnectRequest(request)
	if err != nil {
		return nil, err
	}
	return CheckSessionFromGeneralParams(generalParams)
}

func CheckSessionFromHeaders(headers http.Header) (*auth.SessionData, error) {
	generalParams, err := GeneralParamsFromHeaders(headers)
	if err != nil {
		return nil, err
	}
	return CheckSessionFromGeneralParams(generalParams)
}

func CheckSessionFromGeneralParams(generalParams GeneralParams) (*auth.SessionData, error) {
	if generalParams.Session == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New(ErrorMessages.Unauthorized))
	}
	return generalParams.Session, nil
}

func checkPlatformByUserAgent(userAgent string) string {
	if userAgent == "" {
		return PlatformWeb
	}

	ua := strings.ToLower(userAgent)

	// --- Detección de iOS ---
	// Patrones comunes de iOS: iphone, ipad, ipod
	// Safari en iOS: "mozilla/5.0 (iphone; cpu iphone os 14_7_1 like mac os x) applewebkit/605.1.15 (khtml, like gecko) version/14.1.2 mobile/15e148 safari/605.1.15"
	// Chrome en iOS (es una App): "mozilla/5.0 (iphone; cpu iphone os 14_7_1 like mac os x) applewebkit/605.1.15 (khtml, like gecko) crios/92.0.4515.159 mobile/15e148 safari/604.1"
	// Firefox en iOS (es una App): "mozilla/5.0 (iphone; cpu iphone os 14_7_1 like mac os x) applewebkit/605.1.15 (khtml, like gecko) fxiOS/36.0  mobile/15e148 safari/605.1.15"
	// App Nativa (ej. con User-Agent personalizado): "MiAppSwift/1.0 (iOS 15.0; iPhone13,2)"
	// App Nativa (ej. con URLSession por defecto, puede ser más genérico o incluir CFNetwork): "AppName/1.0 CFNetwork/1206 Darwin/20.1.0"
	if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") || strings.Contains(ua, "ipod") {
		// Indicadores de Navegadores NO Safari (considerados Apps en el ecosistema iOS)
		if strings.Contains(ua, "crios/") || strings.Contains(ua, "fxios/") {
			return PlatformWeb // Chrome o Firefox en iOS son apps que usan WebKit
		}
		// Indicador de Safari
		if strings.Contains(ua, "safari/") && strings.Contains(ua, "version/") && strings.Contains(ua, "mobile/") {
			// Podría ser una WKWebView que no ha modificado su User Agent,
			// o Safari mismo. Para una distinción más fina, se necesitarían
			// User-Agents personalizados desde la app.
			// Si el User-Agent *no* contiene un identificador de app específico que *tú* controlas,
			// es difícil distinguir una WKWebView de Safari. Asumimos Safari por ahora si cumple el patrón.
			// Si quieres ser más estricto, y si tus apps con WKWebView SÍ añaden un identificador,
			// entonces podrías marcar esto como IOSApp si NO tiene tu identificador.
			// Ejemplo: Si tu app añade "MiApp/", y no está, entonces es Safari.
			// if strings.Contains(ua, "miAppEspecifica/") { return IOSApp }
			return PlatformWeb
		}
		// Si es iOS pero no encaja en el patrón de Safari y no es otro navegador conocido,
		// es probable que sea una App (petición directa o WebView con User-Agent modificado).
		// También, si ves CFNetwork, es una app.
		if strings.Contains(ua, "cfnetwork") {
			return PlatformIOS
		}
		// Heurística: si no tiene "mozilla" y es iOS, probablemente una app con User-Agent muy simple
		if !strings.Contains(ua, "mozilla") && (strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") || strings.Contains(ua, "ipod")) {
			return PlatformIOS
		}
		// Fallback para iOS: si contiene identificadores de iOS pero no se ajusta a Safari, se asume App.
		// Esto es amplio; User-Agents personalizados en la app son la mejor manera de estar seguros.
		return PlatformIOS
	}

	// --- Detección de Android ---
	// Patrones comunes de Android: "android"
	// Chrome en Android: "mozilla/5.0 (linux; android 11; sm-g975f) applewebkit/537.36 (khtml, like gecko) chrome/92.0.4515.159 mobile safari/537.36"
	// WebView en Android App: "mozilla/5.0 (linux; android 10; sm-a205u) applewebkit/537.36 (khtml, like gecko) version/4.0 chrome/92.0.4515.159 mobile safari/537.36 myapp/1.0 example_webview; wv)" -> el "; wv)" es clave
	// App Nativa Kotlin con OkHttp (por defecto): "okhttp/4.9.1"
	// App Nativa Kotlin con User-Agent personalizado: "MiAppKotlin/2.1 (Android 11; Pixel 5)"
	if strings.Contains(ua, "android") {
		// Indicador de WebView en Android
		if strings.Contains(ua, "; wv)") || strings.Contains(ua, "webview") {
			return PlatformAndroid
		}
		// Indicador de navegador Chrome o similar en Android
		if strings.Contains(ua, "chrome/") && strings.Contains(ua, "mobile") {
			return PlatformWeb
		}
		// Indicador de Firefox en Android
		if (strings.Contains(ua, "firefox/") || strings.Contains(ua, "gecko/")) && strings.Contains(ua, "mobile") {
			return PlatformWeb
		}
		// User-Agent de librerías comunes como OkHttp (sin la parafernalia de navegador)
		// Usamos una expresión regular simple para `okhttp/numero.numero.numero`
		// También si no tiene "mozilla" pero sí "android", es probable una app
		isOkHttp, _ := regexp.MatchString(`okhttp\/[0-9]+\.[0-9]+(\.[0-9]+)?`, ua)
		if isOkHttp || !strings.Contains(ua, "mozilla") {
			return PlatformAndroid
		}

		// Si es Android y no hemos podido clasificarlo como app o navegador específico,
		// pero parece un User-Agent de navegador (contiene "mozilla" y "applewebkit"),
		// lo clasificamos como Navegador Android genérico.
		if strings.Contains(ua, "mozilla") && strings.Contains(ua, "applewebkit") {
			return PlatformWeb
		}
		// Fallback para Android: si solo dice "android" y no encaja en otros, se asume App.
		return PlatformAndroid
	}

	// Si no es Android ni iOS, asumimos Web Desktop/Otro
	// Aquí podrías añadir más lógica para detectar otros OS móviles o tipos de cliente si es necesario.
	// Por ejemplo, navegadores en KaiOS, Windows Phone (aunque ya obsoleto), etc.
	// O distinguir entre navegadores de escritorio.
	// "mozilla/..." es común en navegadores de escritorio.
	if strings.Contains(ua, "windows nt") || strings.Contains(ua, "macintosh") || strings.Contains(ua, "x11;") {
		return PlatformWeb
	}

	// Si no se pudo identificar claramente, pero contiene "mozilla" (típico de navegadores)
	if strings.Contains(ua, "mozilla") {
		return PlatformWeb // O un "Web (Otro Navegador)" más genérico
	}

	// Si es algo como "curl/7.68.0" o "PostmanRuntime/7.29.0"
	if strings.Contains(ua, "curl/") || strings.Contains(ua, "postmanruntime/") {
		// Podrías tener una categoría "HerramientaAPI" o similar
		return PlatformUnknown // O una categoría específica si lo deseas
	}

	return PlatformUnknown
}
