const SearchEndpoint = "http://localhost:3000/search?q="

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
        const listItem = `
          <li class="result-content--item">
            <span class="result-content--item-doc">${doc.doc}</span>
            <span class="result-content--item-carrot"></span>
            <span class="result-content--item-rel">${doc.tfidf}</span>
          </li>
        `
        docItems.push(listItem);
      }

      fragment.innerHTML = docItems.join("");
    }

    this.#contentHandle.appendChild(fragment.content);
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
