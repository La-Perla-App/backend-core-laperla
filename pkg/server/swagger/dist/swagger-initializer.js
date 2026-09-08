window.onload = function() {
  let specPatched = false;

  const ui = SwaggerUIBundle({
    url: "./openapi.yaml",
    dom_id: "#swagger-ui",
    deepLinking: true,
    requestInterceptor: (req) => {
      const auth = window.ui && window.ui.authSelectors.authorized().toJS();
      if (auth && auth.bearerAuth && auth.bearerAuth.value) {
        req.headers.Authorization = `Bearer ${auth.bearerAuth.value}`;
      }
      return req;
    },
    onComplete: function() {
      if (specPatched) return;
      let spec = ui.specSelectors.specJson().toJS();
      if (!spec.components) spec.components = {};
      spec.components.securitySchemes = {
        bearerAuth: {
          type: "http",
          scheme: "bearer",
          bearerFormat: "JWT"
        }
      };
      if (spec.paths) {
        Object.keys(spec.paths).forEach(path => {
          Object.keys(spec.paths[path]).forEach(method => {
            spec.paths[path][method].security = [{ bearerAuth: [] }, {}];
          });
        });
      }
      specPatched = true;
      ui.specActions.updateSpec(JSON.stringify(spec));
    },
    supportedSubmitMethods: ["get", "post", "put", "delete", "patch", "options", "head"],
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
    plugins: [SwaggerUIBundle.plugins.DownloadUrl],
    layout: "BaseLayout",
  });

  window.ui = ui;
};
