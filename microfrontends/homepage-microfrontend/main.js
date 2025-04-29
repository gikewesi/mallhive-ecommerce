// Fix for Cart microfrontend navigation
document.getElementById('cartLink').onclick = function(e) {
  e.preventDefault();
  alert('Navigating to Shopping Cart microfrontend...');
  // Example: window.location.href = 'https://cart.mallhive.com';
};

// Fix for User Profile microfrontend navigation
document.getElementById('userProfileLink').onclick = function(e) {
  e.preventDefault();
  alert('Navigating to User Profile microfrontend...');
  // Example: window.location.href = 'https://profile.mallhive.com';
};

// Fix for Help (external or internal)
document.getElementById('helpLink').onclick = function(e) {
  e.preventDefault();
  alert('Navigating to Help Center...');
  // Example: window.location.href = '/help';
};

// Fix for Product microfrontend for categories
document.querySelectorAll('.category-link').forEach(link => {
  link.onclick = function(e) {
    e.preventDefault();
    const category = link.getAttribute('data-category');
    alert('Navigating to Product microfrontend, category: ' + category);
    // Example: window.location.href = `https://products.mallhive.com/${category}`;
  };
});

// Fix for Product microfrontend for search
function goToProductMicrofrontend() {
  const query = document.getElementById('searchInput').value;
  if (query) {
    alert('Searching "' + query + '" in Product microfrontend...');
    // Example: window.location.href = `https://products.mallhive.com/search?q=${encodeURIComponent(query)}`;
  }
}

// Fix for Product microfrontend for product card
document.querySelectorAll('.product-card').forEach(card => {
  card.onclick = function() {
    const productId = card.getAttribute('data-product-id');
    alert('Navigating to Product microfrontend, product ID: ' + productId);
    // Example: window.location.href = `https://products.mallhive.com/product/${productId}`;
  };
});
