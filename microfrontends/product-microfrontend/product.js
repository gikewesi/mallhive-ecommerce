// product.js

// Product Microfrontend Communication Logic
const PRODUCT_EVENTS = {
  PRODUCT_VIEWED: 'PRODUCT_VIEWED',
  ADD_TO_CART: 'ADD_TO_CART',
  RELATED_PRODUCTS_LOADED: 'RELATED_PRODUCTS_LOADED',
  PRODUCT_SEARCH: 'PRODUCT_SEARCH'
};

let currentProductId = null;

function initializeProductMicrofrontend() {
  window.addEventListener('PRODUCT_NAVIGATE', handleProductNavigation);
  window.addEventListener('PRODUCT_SEARCH', handleProductSearch);

  // Get product ID from URL
  const urlParams = new URLSearchParams(window.location.search);
  currentProductId = urlParams.get('id');

  if (currentProductId) {
    loadProductDetails(currentProductId);
    loadRelatedProducts(currentProductId);
  }

  setupEventListeners();
}

// Handle Product Navigation Events
function handleProductNavigation(event) {
  const productId = event.detail.productId;
  window.history.pushState({}, '', `?id=${productId}`);
  currentProductId = productId;
  loadProductDetails(productId);
  loadRelatedProducts(productId);
}

// Handle Product Search Events
function handleProductSearch(event) {
  const searchQuery = event.detail.query;
  window.location.href = `/search?q=${encodeURIComponent(searchQuery)}`;
}

// Load Product Details from Microservice
async function loadProductDetails(productId) {
  try {
    const response = await fetch(`http://localhost:4100/api/v1/products/${productId}`);
    const productData = await response.json();
    updateProductUI(productData);
    window.dispatchEvent(new CustomEvent(PRODUCT_EVENTS.PRODUCT_VIEWED, { detail: productData }));
  } catch (error) {
    console.error('Error loading product:', error);
    showError('Failed to load product details');
  }
}

// Load Related Products
async function loadRelatedProducts(productId) {
  try {
    const response = await fetch(`http://localhost:4100/api/v1/products/${productId}/related`);
    const relatedProducts = await response.json();
    renderRelatedProducts(relatedProducts);
    window.dispatchEvent(new CustomEvent(PRODUCT_EVENTS.RELATED_PRODUCTS_LOADED, { detail: relatedProducts }));
  } catch (error) {
    console.error('Error loading related products:', error);
  }
}

// Update Product UI
function updateProductUI(product) {
  document.title = `${product.name} | Mallhive`;
  document.getElementById('productTitle').textContent = product.name;
  document.getElementById('productPrice').textContent = `$${product.price.toLocaleString()}`;
  document.getElementById('originalPrice').textContent = product.originalPrice ? `$${product.originalPrice.toLocaleString()}` : '';
  document.getElementById('discountPercentage').textContent = product.discountPercentage ? `-${product.discountPercentage}%` : '';

  const mainImage = document.getElementById('mainProductImage');
  mainImage.src = product.images[0];

  const thumbnailContainer = document.getElementById('thumbnailContainer');
  thumbnailContainer.innerHTML = product.images.map(img => `
    <img src="${img}" onclick="changeMainImage('${img}')">
  `).join('');

  const featuresList = document.getElementById('productFeatures');
  featuresList.innerHTML = product.features.map(f => `<li>${f}</li>`).join('');

  const specsTable = document.getElementById('specsTable');
  specsTable.innerHTML = Object.entries(product.specifications).map(([key, value]) => `
    <tr>
      <td>${key}</td>
      <td>${value}</td>
    </tr>
  `).join('');
}

// Render Related Products
function renderRelatedProducts(products) {
  const relatedContainer = document.getElementById('relatedProducts');
  relatedContainer.innerHTML = products.map(product => `
    <div class="related-product" onclick="navigateToProduct('${product.id}')">
      <img src="${product.thumbnail}">
      <p class="related-title">${product.name}</p>
      <p class="related-price">$${product.price.toLocaleString()}</p>
    </div>
  `).join('');
}

// Event Listeners
function setupEventListeners() {
  document.getElementById('addToCartBtn').addEventListener('click', () => {
    const quantity = parseInt(document.getElementById('productQty').value);

    // Trigger the event to the Cart MF
    window.dispatchEvent(new CustomEvent(PRODUCT_EVENTS.ADD_TO_CART, {
      detail: {
        productId: currentProductId,
        quantity: quantity
      }
    }));

    // Navigate to the Cart Microfrontend
    window.location.replace('http://localhost:3300/index.html');
  });

  document.getElementById('buyNowBtn').addEventListener('click', () => {
    window.location.replace('http://localhost:3400/index.html');
  });

  document.getElementById('userProfileLink').addEventListener('click', () => {
    window.location.replace('http://localhost:3500/index.html');
  });

  document.getElementById('cartIcon').addEventListener('click', () => {
    window.location.replace('http://localhost:3300/index.html');
  });

  document.getElementById('increaseQty').addEventListener('click', () => {
    const qtyInput = document.getElementById('productQty');
    qtyInput.value = parseInt(qtyInput.value) + 1;
  });

  document.getElementById('decreaseQty').addEventListener('click', () => {
    const qtyInput = document.getElementById('productQty');
    if (qtyInput.value > 1) {
      qtyInput.value = parseInt(qtyInput.value) - 1;
    }
  });
}

// Navigation Functions
function navigateToProduct(productId) {
  window.dispatchEvent(new CustomEvent('PRODUCT_NAVIGATE', {
    detail: { productId }
  }));
}

function changeMainImage(imgUrl) {
  document.getElementById('mainProductImage').src = imgUrl;
}

// Error Handling
function showError(message) {
  const errorDiv = document.createElement('div');
  errorDiv.className = 'error-message';
  errorDiv.innerHTML = `<i class="fas fa-exclamation-triangle"></i><p>${message}</p>`;
  document.querySelector('.product-details').prepend(errorDiv);
}

// Expose API for other Microfrontends
window.ProductMicrofrontend = {
  navigateToProduct: (productId) => {
    navigateToProduct(productId);
  },
  searchProducts: (query) => {
    window.dispatchEvent(new CustomEvent(PRODUCT_EVENTS.PRODUCT_SEARCH, {
      detail: { query }
    }));
  }
};

// Initialize when DOM loaded
document.addEventListener('DOMContentLoaded', initializeProductMicrofrontend);
