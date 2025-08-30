let articleElements = document.querySelectorAll('article[data-published]');
let dateElement = document.getElementById('filter-by-date');

function getQueryParam(name) {
    const urlParams = new URLSearchParams(window.location.search);
    return urlParams.get(name);
}

function updateURLWithDate(curDateParam) {
    const url = new URL(window.location.href);
    if (curDateParam) {
        url.searchParams.set('curDate', curDateParam);
    } else {
        url.searchParams.delete('curDate');
    }
    window.history.pushState({}, '', url.toString());
}

function updateNavLinks(curDateParam) {
    document.querySelectorAll('nav a').forEach(link => {
        const url = new URL(link.href.split('?')[0], window.location.origin);
        if (curDateParam) {
            url.searchParams.set('curDate', curDateParam);
        } else {
            url.searchParams.delete('curDate');
        }
        link.href = url.toString();
    });
}

function filterArticles(event) {
    const selectedDate = event.target.value; // Use event.target to get the correct element
    articleElements.forEach(article => {
        const articleDate = article.getAttribute('data-published').split('T')[0];
        if (selectedDate === '') {
            article.classList.remove('hidden');
        } else if (articleDate === selectedDate) {
            article.classList.remove('hidden');
        } else {
            article.classList.add('hidden');
        }
    });
    // Update the URL parameters and navigation links
    updateURLWithDate(selectedDate);
    updateNavLinks(selectedDate);
}

document.addEventListener('DOMContentLoaded', function() {
  articleElements = document.querySelectorAll('article[data-published]');
  dateElement = document.getElementById('filter-by-date');
  dateElement.addEventListener('change', filterArticles);
  let minDate = new Date();
  let maxDate = new Date();
  // Set the default values
  dateElement.setAttribute('min', minDate.toISOString().split('T')[0]);
  dateElement.setAttribute('max', maxDate.toISOString().split('T')[0]);
  // Find and update the min/max dates in our article list.
  articleElements.forEach(article => {
    const curDate = new Date(article.getAttribute('data-published'));
    if (curDate >= maxDate) {
        dateElement.setAttribute('max', curDate.toISOString().split('T')[0]);
        maxDate = curDate;
    }
    if (curDate <= minDate) {
        dateElement.setAttribute('min', curDate.toISOString().split('T')[0]);
        minDate = curDate;
    }
  });

  // Check for curDate parameter in the URL
  const curDateParam = getQueryParam('curDate');
  if (curDateParam) {
      dateElement.value = curDateParam;
      // Simulate triggering the change event directly
      filterArticles({ target: dateElement });
  }
});

export {}; // This makes the file a module
