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
      await navigator.clipboard.writeText(text)
    })
  }
};
