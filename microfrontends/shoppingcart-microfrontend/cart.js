// cart.js

// Cart Microfrontend Communication Logic
const CART_EVENTS = {
  CART_UPDATED: 'CART_UPDATED',
  REMOVE_FROM_CART: 'REMOVE_FROM_CART'
};

let currentCart = {};

// Initialize Cart Microfrontend
function initializeCartMicrofrontend() {
  // Get cart ID from localStorage or URL
  loadCartData();

  // Setup event listeners
  setupEventListeners();
}

// Load Cart Data
function loadCartData() {
  // Assuming the cart data is stored in memory or localStorage
  const storedCart = localStorage.getItem('cart');
  currentCart = storedCart ? JSON.parse(storedCart) : {};

  updateCartUI();
}

// Update Cart UI
function updateCartUI() {
  const cartContainer = document.getElementById('cartItems');
  cartContainer.innerHTML = ''; // Clear the cart

  // Display the items in the cart
  Object.entries(currentCart).forEach(([productId, product]) => {
    const cartItem = document.createElement('div');
    cartItem.classList.add('cart-item');

    cartItem.innerHTML = `
      <img src="${product.image}" alt="${product.name}" class="cart-item-image" />
      <div class="cart-item-details">
        <p class="cart-item-name">${product.name}</p>
        <p class="cart-item-price">$${product.price.toLocaleString()}</p>
        <p class="cart-item-quantity">Quantity: ${product.quantity}</p>
        <button class="remove-from-cart" data-product-id="${productId}">Remove</button>
      </div>
    `;

    cartContainer.appendChild(cartItem);
  });

  // Dispatch the event to notify the cart update
  window.dispatchEvent(new CustomEvent(CART_EVENTS.CART_UPDATED, { detail: currentCart }));
}

// Event Listeners
function setupEventListeners() {
  document.getElementById('checkoutBtn').addEventListener('click', () => {
    window.location.replace('http://localhost:3400/checkout.html');
  });

  document.getElementById('clearCartBtn').addEventListener('click', () => {
    clearCart();
  });

  // Listen for remove button clicks
  document.getElementById('cartItems').addEventListener('click', (event) => {
    if (event.target.classList.contains('remove-from-cart')) {
      const productId = event.target.getAttribute('data-product-id');
      removeFromCart(productId);
    }
  });
}

// Add product to cart
function addToCart(productId, quantity) {
  // Check if product already exists in the cart
  if (currentCart[productId]) {
    currentCart[productId].quantity += quantity; // Update quantity if already in cart
  } else {
    // Add new product to the cart
    currentCart[productId] = { ...productData, quantity };
  }

  saveCartData();
  updateCartUI();
}

// Remove product from cart
function removeFromCart(productId) {
  delete currentCart[productId];
  saveCartData();
  updateCartUI();
}

// Clear the cart
function clearCart() {
  currentCart = {};
  saveCartData();
  updateCartUI();
}

// Save cart data to localStorage (or a database if needed)
function saveCartData() {
  localStorage.setItem('cart', JSON.stringify(currentCart));
}

// Expose API for other Microfrontends
window.CartMicrofrontend = {
  addToCart: (productId, quantity) => {
    addToCart(productId, quantity);
  },
  removeFromCart: (productId) => {
    removeFromCart(productId);
  },
  clearCart: () => {
    clearCart();
  }
};

// Initialize when DOM loaded
document.addEventListener('DOMContentLoaded', initializeCartMicrofrontend);
