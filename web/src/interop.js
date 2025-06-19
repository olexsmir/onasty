import "./styles.css";

export const flags = (_) => {
  return {
    access_token: JSON.parse(window.localStorage.access_token || "null"),
    refresh_token: JSON.parse(window.localStorage.refresh_token || "null"),
  }
}

export const onReady = ({ app }) => {
  if (app.ports?.sendToLocalStorage) {
    app.ports.sendToLocalStorage.subscribe(({ key, value }) => {
      window.localStorage[key] = JSON.stringify(value);
    })
  }
}
