import { AppContextProvider } from "./views/context/appContextProvider.jsx";
import { createRoot } from "react-dom/client";
import { StrictMode } from "react";
import App from "./App.jsx";

createRoot(document.getElementById("root")).render(
  <StrictMode>
    <AppContextProvider>
      <App />
    </AppContextProvider>
  </StrictMode>
);
