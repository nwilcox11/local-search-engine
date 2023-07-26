### Resources

Inspired by [tsoding-search-engine](https://www.youtube.com/watch?v=hm5xOJiVEeg&list=WL&index=37)

1) **[TF-IDF](https://en.wikipedia.org/wiki/Tf%E2%80%93idf)** short for *term frequency–inverse document frequency*,
is a numerical statistic that is intended to reflect how important a word is to a document in a collection.

    ##### Term frequency
    - The weight of a term that occurs in a document is simply proportional to the term frequency.
    - There are various other ways to represent term frequency. We are going to use raw count for now.

    ```
    C = raw count of term in document.
    tf(term, d) = C
    ```

    ##### Inverse document frequency

    - Because a term like "The" is so common, term frequency will tend to incorrectly
    emphasize document swhich happen to use the word "The" more frequently, without
    giving enough weight to to more meaningful words in the search term. This means
    that "The" is not a good keyword to distinguish relavant and non-relavant documents.

    - *Inverse document frequency* factor is included to diminish the weight of terms
    that occur very frequently in the document set, and increases the weight of terms
    that occur rarely.

    - Measure of how much information the word provides. Is it common or rare across
    all the documents.

    - logarithmaically scaled inverse fraction of the documents that contain the word.
        - Divide the total number of documents by the number of documents containing the
        term, and then taking the log of that quotient.

    ```
    N = |D|: Total number of docs.
    D = 1 + D: Number of docs that contain term.

    idf(term, D) = log(N/D)
    ```

    ##### Tf-idf

    ```
    tfidf(t, d, D) = tf(t, d) * idf(t, D)
    ```

2) **[Crafting Interpreters](https://github.com/munificent/craftinginterpreters/tree/master)** The book we are going use in the search engine ranking.
