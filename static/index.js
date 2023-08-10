const SearchEndpoint = "http://localhost:3000/search?q="
const N = 78;
const P = 80;
const S = 83;
const QE = 191; // ?

function processFloat(float) {
  return float.toFixed(3);
}

function processDocTitle(doc) {
  const removeDomain = doc.split("/");
  const removeFiletype = removeDomain[1].split(".");
  return removeFiletype[0].replaceAll("-", " ")
    .split(" ")
    .map(word => word.charAt(0).toUpperCase() + word.substring(1)).join(" ");
}

class App {
  #formHandle = document.getElementById("search-form");
  #inputHandle = document.getElementById("search");
  #contentHandle = document.getElementById("result-content");
  #bodyHandle = document.getElementById("app");
  #mainHandle = document.getElementById("main");

  #content = undefined;
  #anchors = undefined;

  #state = {
    navigatedTo: 0,
  };

  constructor() {
    this.#bindEventListeners();
  }

  bootstrap() {
    this.#formHandle.addEventListener("submit", this.doSearch);
    window.addEventListener("keydown", (e) => { this.keyboardAction(e) });
  }

  async #doSearch(e) {
    e.preventDefault();

    const query = encodeURI(this.#inputHandle.value);
    if (!query) return;

    try {
      const resp = await fetch(`${SearchEndpoint}${query}`);

      // TODO: Currently the server returns 204 on empty search results
      // Lets change this to something useful.
      if (resp.status === 204) return;

      const jsonResp = await resp.json();

      if (this.#content) {
        this.#clearResultList();
      }

      this.#renderResultList(jsonResp);
      this.#content = jsonResp;
      this.#anchors = Array.from(this.#contentHandle.childNodes).filter(node => node.nodeName === "A");
    } catch (err) {
      console.log(err)
    }
  }

  #bindEventListeners() {
    this.doSearch = this.#doSearch.bind(this);
    this.keyboardAction = this.#keyboardAction.bind(this);
  }

  #keyboardAction(e) {
    if (this.#content) {
      this.#navigation(e);
    }

    if (e.ctrlKey && e.keyCode === S) {
      this.#inputHandle.focus();
    }

    if (document.activeElement !== this.#inputHandle && (e.shiftKey && e.keyCode === QE)) {
      this.#renderDialog();
    }
  }

  #navigation(e) {
    const navigationStart = document.activeElement === this.#bodyHandle ||
      document.activeElement === this.#inputHandle

    if (e.ctrlKey && e.keyCode === N) {
      if (navigationStart) {
        this.#state.navigatedTo = 0;
        this.#focusResultAnchor();
      } else {
        this.#state.navigatedTo = (this.#state.navigatedTo + 1) % this.#anchors.length;
        this.#focusResultAnchor();
      }
    }

    if (e.ctrlKey && e.keyCode === P) {
      if (navigationStart) {
        this.#state.navigatedTo = this.#anchors.length - 1;
        this.#anchors[this.#state.navigatedTo].focus();
        this.#focusResultAnchor();
      } else {
        this.#state.navigatedTo = (this.#state.navigatedTo - 1 + this.#anchors.length) % this.#anchors.length;
        this.#focusResultAnchor();
      }
    }
  }

  #focusResultAnchor() {
    this.#anchors.length && this.#anchors[this.#state.navigatedTo].focus();
  }

  #renderResultList(content) {
    if (!content) return;
    const docItems = [];
    const fragment = document.createElement("template");

    for (const [_, docs] of Object.entries(content)) {
      for (const doc of docs) {
        docItems.push(this.#resultCard(doc));
      }

      fragment.innerHTML = docItems.join("");
    }

    this.#contentHandle.appendChild(fragment.content);
  }

  #renderDialog() {
    const helpMenu = document.getElementById("keyboard-shortcuts");
    if (helpMenu) {
      this.#mainHandle.removeChild(helpMenu);
    } else {
      const template = document.createElement("template")
      template.innerHTML = this.#dialog();
      this.#mainHandle.appendChild(template.content);
    }
  }

  #resultCard(doc) {
    return `
      <a href=https://${doc.doc} target=_blank>
        <li class="result-content--card">
          <div class="result-content-group">
            <span class="result-content--card-doc">${processDocTitle(doc.doc)}</span>
            <p class="result-content--card-meta">${doc.meta}</p>
          </div>
          <div class="chip-group">
            ${this.#chip(processFloat(doc.idf), "idf")}
            ${this.#chip(processFloat(doc.tfidf), "tfidf")}
          </div>
        </li>
      </a>
    `
  }

  #chip(text, kind) {
    return `
      <span class="chip">
        <span class="chip-kind ${kind}"></span>
        <span class="chip-text">${text}</span>
      </span>
    `
  }

  #dialog() {
    return `
      <dialog id="keyboard-shortcuts">
        <span class="dialog-title">
          Keyboard Shortcuts
        </span>
        <ul class="dialog-content">
          <div class="dialog-content-group">
            <span class="dialog-content-header">Navigation</span>
            <li class="dialog-content-item">
              <span class="dialog-content-item--label">Next result</span>
              <span class="dialog-content-item--key">
                <kbd>ctrl</kbd>
                <kbd>n</kbd>
              </span>
            </li>
            <li class="dialog-content-item">
              <span class="dialog-content-item--label">Previous result</span>
              <span class="dialog-content-item--key">
                <kbd>ctrl</kbd>
                <kbd>p</kbd>
              </span>
            </li>
          </div>
          <div class="dialog-content-group">
            <span class="dialog-content-header">Action</span>
            <li class="dialog-content-item">
              <span class="dialog-content-item--label">Focus search input</span>
              <span class="dialog-content-item--key">
                <kbd>ctrl</kbd>
                <kbd>s</kbd>
              </span>
            </li>
            <li class="dialog-content-item">
              <span class="dialog-content-item--label">Open help center</span>
              <span class="dialog-content-item--key">
                <kbd>?</kbd>
              </span>
            </li>
          </div>
        </ul>
      </dialog>
    `
  }

  #clearResultList() {
    while (this.#contentHandle.firstChild) {
      this.#contentHandle.removeChild(this.#contentHandle.firstChild);
    }
    this.#anchors = undefined;
    this.#content = undefined;
  }
}

(function main() {
  new App().bootstrap();
})();
