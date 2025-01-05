function Ergon() {
  this.replaceChildren = false;
  this.children = undefined;
  this.element = undefined;
  this.parent = undefined;
  this.type = undefined;
  this.props = {};
}

// Metodo per creare un elemento DOM
Ergon.prototype.createElement = function () {
  const element = document.createElement(this.tagName);

  // Impostare le proprietà
  for (const [key, value] of Object.entries(this.props)) {
    if (key.startsWith("on") && typeof value === "function") {
      // Event listeners (es. onClick)
      element.addEventListener(key.slice(2).toLowerCase(), value);
    } else {
      // Attributi normali (es. href, class, id)
      element.setAttribute(key, value);
    }
  }

  // Aggiungere i figli (se presenti)
  this.children.forEach((child) => {
    if (typeof child === "string" || typeof child === "number") {
      element.appendChild(document.createTextNode(child));
    } else if (child instanceof Node) {
      element.appendChild(child);
    }
  });

  this.element = element;
};

// Metodo per renderizzare un elemento nel contenitore
Ergon.prototype.renderElement = function () {
  const container = document.getElementById(this.parent);
  if (!container) {
    console.error(`Elemento con id "${this.parent}" non trovato.`);
    return;
  }

  if (this.replaceChildren) {
    // Rimuovere tutti i figli esistenti
    container.innerHTML = "";
  }

  // Aggiungere il nuovo elemento
  container.appendChild(this.element);
};

Ergon.prototype.Render = function () {
  this.createElement();
  this.renderElement();
};
