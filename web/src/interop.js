import "./styles.css";

export const flags = ({ env }) => {
  return {
    access_token: JSON.parse(window.localStorage.access_token || "null"),
    refresh_token: JSON.parse(window.localStorage.refresh_token || "null"),
    app_url: env.FRONTEND_URL || "http://localhost:3000",
  };
};

export const onReady = ({ app }) => {
  if (app.ports?.sendToLocalStorage) {
    app.ports.sendToLocalStorage.subscribe(({ key, value }) => {
      window.localStorage[key] = JSON.stringify(value);
    });
  }

  if (app.ports?.sendToClipboard) {
    app.ports.sendToClipboard.subscribe(async (text) => {
      try {
        await navigator.clipboard.writeText(text);
      } catch (error) {
        console.error("Failed to write to clipboard:", error);
      }
    });
  }

  if (app.ports?.confirmRequest && app.ports?.confirmResponse) {
    app.ports.confirmRequest.subscribe(msg => {
      const res = window.confirm(msg);
      app.ports.confirmResponse.send(res);
    });
  }
};
