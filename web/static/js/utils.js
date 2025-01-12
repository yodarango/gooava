// Takes in a form event and maps each field to a json object
function formDataToJson(event) {
  event.preventDefault();
  const form = event.target.closest("form");
  const formData = new FormData(form);
  const data = {};

  Array.from(form.elements).forEach((element) => {
    if (element.name) {
      if (element.type === "checkbox") {
        data[element.name] = element.checked;
      } else if (element.type === "radio") {
        if (element.checked) {
          data[element.name] = element.value;
        }
      } else if (element.type === "range") {
        data[element.name] = Number(formData.get(element.name) || "1");
      } else {
        data[element.name] = formData.get(element.name) || "";
      }
    }
  });

  return { path: form.action, data };
}

// Sends a post request and returns the response in JSON format
async function postRequest(url, payload) {
  try {
    // Invia la richiesta POST al server
    const response = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload), // Convertire il payload in JSON
    });

    // Controlla se la risposta HTTP è valida
    if (!response.ok) {
      throw new Error(`HTTP error! Status: ${response.status}`);
    }

    // Parsa la risposta come JSON
    const responseData = await response.json();

    // Verifica che la risposta contenga il formato previsto
    if (
      typeof responseData.code !== "number" ||
      !("data" in responseData) ||
      typeof responseData.success !== "boolean"
    ) {
      throw new Error("Formato di risposta non valido");
    }
    // Ritorna la risposta
    return responseData;
  } catch (error) {
    console.error("Errore nella richiesta POST:", error);
    throw error; // Rilancia l'errore per la gestione
  }
}

// creates a toast message of variant success, danger, warning, or info
// at the duration specified
function createToast({
  message,
  duration = 5,
  autoClose = true,
  variant = "info",
  onAutoClose,
}) {
  // Creare il contenitore del toast
  const toast = document.createElement("div");
  toast.className = `toast toast-${variant}`; // Aggiungere la variante come classe
  toast.innerHTML = `
  <div class="d-flex align-items-center justify-content-start gap-2">
    <ion-icon name="checkmark-circle"></ion-icon>
    <p>${message}</p>
  </div>

  `;

  // Barra di progresso
  if (autoClose) {
    const progressBar = document.createElement("div");
    progressBar.className = "toast-progress";
    toast.appendChild(progressBar);

    // Animazione della barra di progresso
    setTimeout(() => {
      progressBar.style.transition = `width ${duration}s linear`;
      progressBar.style.width = "0";
    }, 10);
  }

  // Pulsante di chiusura (se autoClose è falso)
  if (!autoClose) {
    const closeButton = document.createElement("button");
    closeButton.className = "toast-close";
    closeButton.textContent = "Close";
    closeButton.addEventListener("click", () => {
      document.body.removeChild(toast);
    });
    toast.appendChild(closeButton);
  }

  // Aggiungere il toast al body
  document.body.appendChild(toast);

  // Rimuovere automaticamente il toast dopo il tempo specificato
  if (autoClose) {
    setTimeout(() => {
      if (toast.parentElement) {
        document.body.removeChild(toast);
        if (onAutoClose) {
          onAutoClose();
        }
      }
    }, duration * 1000);
  }
}
