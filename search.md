---
title : "Search"
---

<noscript>JavaScript must be enabled for search to work</noscript>

# Search Robert's blog posts and selected projects

<link href="/pagefind/pagefind-component-ui.css" rel="stylesheet">
<script src="/pagefind/pagefind-component-ui.js" type="module"></script>

<pagefind-input placeholder="Search…"></pagefind-input>
<pagefind-summary></pagefind-summary>
<pagefind-results></pagefind-results>

<script type="module">
  // Wait for the pagefind web components to register before using the API
  await customElements.whenDefined('pagefind-input');

  const instance = window.PagefindComponents.getInstanceManager().getInstance('default');

  // Pre-populate from URL query parameter (supports osd.xml OpenSearch integration)
  const q = new URLSearchParams(location.search).get('q');
  if (q) {
    instance.triggerSearch(q);
  }

  // Keep URL in sync with the current search term so results are bookmarkable/shareable
  instance.on('search', (term) => {
    const url = new URL(location.href);
    if (term) {
      url.searchParams.set('q', term);
    } else {
      url.searchParams.delete('q');
    }
    history.replaceState(null, '', url.toString());
  });
</script>
