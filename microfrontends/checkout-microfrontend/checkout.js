/**
 * Checkout Microfrontend
 * 
 * This script handles:
 * 1. Communication with Payment microservice
 * 2. Receiving order data from Shopping Cart microfrontend
 * 3. Managing the checkout flow
 * 4. Processing payment method selection
 */

// Checkout State
let checkoutState = {
    orderId: null,
    customerId: null,
    items: [],
    paymentMethod: 'card_bank_ussd',
    address: {
      name: 'Gini DevOps ',
      street: 'Addision avenue',
      city: 'Dallas - Texas',
      phone: '+1 618-298-9051'
    },
    delivery: {
      method: 'pickup',
      startDate: '06 May',
      endDate: '13 May'
    },
    totals: {
      itemsTotal: 355,
      deliveryFee: 27,
      orderTotal: 358
    }
  };
  
  // API Endpoints for Payment Microservice
  const PAYMENT_API = {
    getPaymentMethods: 'https://api.mallhive.com/payments/methods',
    validatePayment: 'https://api.mallhive.com/payments/validate',
    processPayment: 'https://api.mallhive.com/payments/process',
    applyPromoCode: 'https://api.mallhive.com/payments/promo'
  };
  
  // Event Names for Microfrontend Communication
  const EVENTS = {
    // Events we listen for (incoming)
    CART_CHECKOUT_REQUESTED: 'CART_CHECKOUT_REQUESTED',
    PAYMENT_PROCESSED: 'PAYMENT_PROCESSED',
    ORDER_UPDATED: 'ORDER_UPDATED',
    
    // Events we dispatch (outgoing)
    CHECKOUT_INITIALIZED: 'CHECKOUT_INITIALIZED',
    PAYMENT_METHOD_SELECTED: 'PAYMENT_METHOD_SELECTED',
    PAYMENT_INITIATED: 'PAYMENT_INITIATED',
    ORDER_COMPLETED: 'ORDER_COMPLETED',
    CHECKOUT_ERROR: 'CHECKOUT_ERROR'
  };
  
  /**
   * Initialize checkout with data from Shopping Cart
   */
  function initializeCheckout(cartData) {
    try {
      if (!cartData) {
        // If no cart data is provided, try to get it from localStorage
        const savedCartData = localStorage.getItem('mallhive_cart');
        if (savedCartData) {
          cartData = JSON.parse(savedCartData);
        } else {
          throw new Error('No cart data available');
        }
      }
      
      // Update checkout state with cart data
      checkoutState.items = cartData.items || [];
      checkoutState.totals.itemsTotal = cartData.subtotal || 0;
      checkoutState.totals.orderTotal = checkoutState.totals.itemsTotal + checkoutState.totals.deliveryFee;
      
      // Generate a temporary order ID if not provided
      checkoutState.orderId = cartData.orderId || `ORD-${Date.now()}`;
      
      // Update the UI with the order details
      updateOrderSummary();
      
      // Fetch available payment methods from the Payment microservice
      fetchPaymentMethods();
      
      // Notify other microfrontends that checkout has been initialized
      const event = new CustomEvent(EVENTS.CHECKOUT_INITIALIZED, {
        detail: {
          orderId: checkoutState.orderId,
          totals: checkoutState.totals
        }
      });
      window.dispatchEvent(event);
      
    } catch (error) {
      console.error('Error initializing checkout:', error);
      notifyError('Failed to initialize checkout. Please try again.');
    }
  }
  
  /**
   * Fetch available payment methods from Payment microservice
   */
  async function fetchPaymentMethods() {
    try {
      // In a real application, this would be an actual API call
      // const response = await fetch(PAYMENT_API.getPaymentMethods);
      // const methods = await response.json();
      
      // For demo purposes, we'll use the methods already in the HTML
      // This would typically update the available payment methods dynamically
      
      // Enable/disable payment methods based on order value and other factors
      const orderValue = checkoutState.totals.orderTotal;
      
      // In a real scenario, these rules would come from the Payment microservice
      if (orderValue < 8500 || orderValue > 450000) {
        document.querySelectorAll('.payment-option.disabled').forEach(option => {
          option.querySelector('input').disabled = true;
        });
      }
      
    } catch (error) {
      console.error('Error fetching payment methods:', error);
      notifyError('Failed to load payment methods. Please try again.');
    }
  }
  
  /**
   * Update the order summary in the UI
   */
  function updateOrderSummary() {
    document.getElementById('items-total').textContent = formatCurrency(checkoutState.totals.itemsTotal);
    document.getElementById('delivery-fees').textContent = formatCurrency(checkoutState.totals.deliveryFee);
    document.getElementById('order-total').textContent = formatCurrency(checkoutState.totals.orderTotal);
    
    // Update delivery dates
    document.getElementById('delivery-start-date').textContent = checkoutState.delivery.startDate;
    document.getElementById('delivery-end-date').textContent = checkoutState.delivery.endDate;
    
    // Update customer information
    document.getElementById('customer-name').textContent = checkoutState.address.name;
    document.getElementById('customer-address-line').textContent = 
      `${checkoutState.address.street} | ${checkoutState.address.city} | ${checkoutState.address.phone}`;
  }
  
  /**
   * Process the selected payment method
   */
  async function processPaymentMethod() {
    try {
      // Get the selected payment method
      const selectedMethod = document.querySelector('input[name="payment"]:checked').value;
      checkoutState.paymentMethod = selectedMethod;
      
      // Notify about payment method selection
      const event = new CustomEvent(EVENTS.PAYMENT_METHOD_SELECTED, {
        detail: {
          orderId: checkoutState.orderId,
          paymentMethod: checkoutState.paymentMethod
        }
      });
      window.dispatchEvent(event);
      
      // Different flow based on payment method
      switch (selectedMethod) {
        case 'card_bank_ussd':
          // Redirect to secure payment page
          window.location.href = `https://pay.mallhive.com?orderId=${checkoutState.orderId}`;
          break;
          
        case 'palmpay':
        case 'opay':
          // Redirect to third-party payment provider
          window.location.href = `https://${selectedMethod}.com/checkout?orderId=${checkoutState.orderId}&amount=${checkoutState.totals.orderTotal}`;
          break;
          
        case 'credpal':
        case 'easybuy':
          // Redirect to BNPL provider
          window.location.href = `https://${selectedMethod}.com/checkout?orderId=${checkoutState.orderId}`;
          break;
          
        case 'installment':
          // Show installment options modal
          showInstallmentOptions();
          break;
          
        default:
          throw new Error('Invalid payment method selected');
      }
      
    } catch (error) {
      console.error('Error processing payment method:', error);
      notifyError('Failed to process payment method. Please try again.');
    }
  }
  
  /**
   * Process promotion code
   */
  async function applyPromoCode() {
    try {
      const promoCode = document.getElementById('promo-code-input').value.trim();
      
      if (!promoCode) {
        alert('Please enter a promotion code');
        return;
      }
      
      // In a real application, this would be an actual API call
      // const response = await fetch(`${PAYMENT_API.applyPromoCode}?code=${promoCode}`);
      // const result = await response.json();
      
      // For demo, simulate a successful promo code application
      setTimeout(() => {
        // Update the checkout state with the discount
        const discount = 5000; // Example discount amount
        checkoutState.totals.discount = discount;
        checkoutState.totals.orderTotal -= discount;
        
        // Update the UI
        updateOrderSummary();
        
        // Add discount line to order summary
        const summaryItems = document.querySelector('.summary-item.total').parentNode;
        const discountItem = document.createElement('div');
        discountItem.className = 'summary-item';
        discountItem.innerHTML = `
          <span>Discount (Promo: ${promoCode})</span>
          <span class="price">-$ ${discount.toLocaleString()}</span>
        `;
        summaryItems.insertBefore(discountItem, document.querySelector('.summary-item.total'));
        
        // Clear the input
        document.getElementById('promo-code-input').value = '';
        
        alert('Promotion code applied successfully!');
      }, 500);
      
    } catch (error) {
      console.error('Error applying promo code:', error);
      alert('Invalid promotion code. Please try again.');
    }
  }
  
  /**
   * Complete the order and process final payment
   */
  async function confirmOrder() {
    try {
      // Check if payment method is selected
      if (!checkoutState.paymentMethod) {
        alert('Please select a payment method');
        return;
      }
      
      // Notify that payment is being initiated
      const event = new CustomEvent(EVENTS.PAYMENT_INITIATED, {
        detail: {
          orderId: checkoutState.orderId,
          paymentMethod: checkoutState.paymentMethod,
          amount: checkoutState.totals.orderTotal
        }
      });
      window.dispatchEvent(event);
      
      // For demo purposes, simulate successful order confirmation
      // In a real app, this would communicate with the Payment microservice
      setTimeout(() => {
        // Redirect to selected payment method handling
        processPaymentMethod();
      }, 500);
      
    } catch (error) {
      console.error('Error confirming order:', error);
      notifyError('Failed to confirm your order. Please try again.');
    }
  }
  
  /**
   * Show installment options modal (simplified for demo)
   */
  function showInstallmentOptions() {
    alert('Installment options will be shown here');
    // In a real application, you would show a modal with installment options
  }
  
  /**
   * Format currency for display
   */
  function formatCurrency(amount) {
    return '$ ' + amount.toLocaleString();
  }
  
  /**
   * Notify about errors
   */
  function notifyError(message) {
    const event = new CustomEvent(EVENTS.CHECKOUT_ERROR, {
      detail: { message }
    });
    
    window.dispatchEvent(event);
    
    // For demo, show alert
    alert(message);
  }
  
  /**
   * Initialize event listeners for user interactions
   */
  function initUIEventListeners() {
    // Payment method selection
    document.querySelectorAll('input[name="payment"]').forEach(radio => {
      radio.addEventListener('change', (e) => {
        // Hide all payment details sections
        document.querySelectorAll('.payment-details').forEach(section => {
          section.style.display = 'none';
        });
        
        // Show details for selected payment method if available
        const methodValue = e.target.value;
        const detailsSection = document.querySelector(`.${methodValue}-details`);
        if (detailsSection) {
          detailsSection.style.display = 'block';
        }
      });
    });
    
    // Confirm payment method button
    document.getElementById('confirm-payment-button').addEventListener('click', processPaymentMethod);
    
    // Apply promo code button
    document.getElementById('apply-promo').addEventListener('click', applyPromoCode);
    
    // Confirm order button
    document.getElementById('confirm-order-button').addEventListener('click', confirmOrder);
    
    // Change buttons for previous steps
    document.querySelectorAll('.change-btn').forEach(button => {
      button.addEventListener('click', (e) => {
        const stepId = e.target.closest('.checkout-step').id;
        
        if (stepId === 'customer-address') {
          // Navigate to address edit
          window.location.href = '/checkout/address';
        } else if (stepId === 'delivery-details') {
          // Navigate to delivery options
          window.location.href = '/checkout/delivery';
        }
      });
    });
  }
  
  /**
   * Initialize event listeners for microfrontend communication
   */
  function initCommunicationEventListeners() {
    // Listen for checkout request from Shopping Cart microfrontend
    window.addEventListener(EVENTS.CART_CHECKOUT_REQUESTED, (event) => {
      const cartData = event.detail;
      initializeCheckout(cartData);
    });
    
    // Listen for payment processed event from Payment microservice
    window.addEventListener(EVENTS.PAYMENT_PROCESSED, (event) => {
      const paymentResult = event.detail;
      
      if (paymentResult.success) {
        // Payment was successful
        const orderCompletedEvent = new CustomEvent(EVENTS.ORDER_COMPLETED, {
          detail: {
            orderId: checkoutState.orderId,
            paymentId: paymentResult.paymentId
          }
        });
        window.dispatchEvent(orderCompletedEvent);
        
        // Redirect to order confirmation page
        window.location.href = `/order-confirmation?orderId=${checkoutState.orderId}`;
      } else {
        // Payment failed
        notifyError(`Payment failed: ${paymentResult.message}`);
      }
    });
    
    // Listen for order updates
    window.addEventListener(EVENTS.ORDER_UPDATED, (event) => {
      const updatedOrder = event.detail;
      
      // Update checkout state with new order data
      Object.assign(checkoutState, updatedOrder);
      
      // Update UI
      updateOrderSummary();
    });
  }
  
  /**
   * Check if we're navigating back from a payment page
   */
  function checkPaymentReturn() {
    const urlParams = new URLSearchParams(window.location.search);
    const paymentStatus = urlParams.get('payment_status');
    
    if (paymentStatus === 'success') {
      // Payment was successful
      const paymentId = urlParams.get('payment_id');
      const orderId = urlParams.get('order_id');
      
      // Notify about successful payment
      const event = new CustomEvent(EVENTS.PAYMENT_PROCESSED, {
        detail: {
          success: true,
          orderId: orderId,
          paymentId: paymentId
        }
      });
      window.dispatchEvent(event);
      
      // Redirect to order confirmation
      window.location.href = `/order-confirmation?orderId=${orderId}`;
    } else if (paymentStatus === 'fail') {
      // Payment failed
      notifyError('Payment was not successful. Please try again.');
    }
  }
  
  /**
   * Initialize checkout microfrontend
   */
  function initCheckout() {
    // Check if we're returning from a payment page
    checkPaymentReturn();
    
    // Initialize UI event listeners
    initUIEventListeners();
    
    // Initialize communication event listeners
    initCommunicationEventListeners();
    
    // Try to initialize checkout with data from URL or localStorage
    const urlParams = new URLSearchParams(window.location.search);
    const orderId = urlParams.get('orderId');
    
    if (orderId) {
      // We have an order ID, try to load order data
      // In a real app, you would fetch this from an Order microservice
      const savedOrderData = localStorage.getItem(`mallhive_order_${orderId}`);
      if (savedOrderData) {
        initializeCheckout(JSON.parse(savedOrderData));
      }
    } else {
      // No order ID, check for cart data
      const savedCartData = localStorage.getItem('mallhive_cart');
      if (savedCartData) {
        initializeCheckout(JSON.parse(savedCartData));
      }
    }
  }
  
  // Initialize when DOM is ready
  document.addEventListener('DOMContentLoaded', initCheckout);
  
  /**
   * Exposed API for other microfrontends to use directly
   */
  window.CheckoutMicrofrontend = {
    startCheckout: initializeCheckout,
    updateOrderData: (orderData) => {
      Object.assign(checkoutState, orderData);
      updateOrderSummary();
    },
    getCheckoutState: () => ({ ...checkoutState }) // Return a copy to prevent direct mutation
  };
  