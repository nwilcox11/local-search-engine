const SearchEndpoint = "http://localhost:3000/search?q="

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
  #content = undefined;

  constructor() {
    this.#bindEventListeners();
  }

  bootstrap() {
    this.#formHandle.addEventListener("submit", this.doSearch);
  }

  async #doSearch(e) {
    e.preventDefault();

    const query = encodeURI(this.#inputHandle.value);

    try {
      const resp = await fetch(`${SearchEndpoint}${query}`);
      const content = await resp.json();

      if (this.#content) {
        this.#clearResultList();
      }

      this.#renderResultList(content);
      this.#content = content;
    } catch (err) {
      console.log(err)
    }
  }

  #bindEventListeners() {
    this.doSearch = this.#doSearch.bind(this);
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

  #clearResultList() {
    while (this.#contentHandle.firstChild) {
      this.#contentHandle.removeChild(this.#contentHandle.firstChild);
    }
  }
}

(function main() {
  new App().bootstrap();
})();
