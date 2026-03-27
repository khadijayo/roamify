(function () {
  function showError(message) {
    var el = document.getElementById("swagger-error");
    if (!el) {
      return;
    }
    el.style.display = "block";
    el.textContent = message;
  }

  function bootSwagger() {
    try {
      if (typeof SwaggerUIBundle !== "function") {
        throw new Error("Swagger UI bundle is not available");
      }

      SwaggerUIBundle({
        url: "openapi.json",
        dom_id: "#swagger-ui",
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis],
        layout: "BaseLayout"
      });
    } catch (err) {
      showError(
        "Swagger UI failed to load.\n" +
          "Reason: " + (err && err.message ? err.message : String(err)) +
          "\n\n" +
          "Open the raw spec at /swagger/openapi.json"
      );
    }
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", bootSwagger);
  } else {
    bootSwagger();
  }
})();
