// UserProfile Microfrontend

const USER_PROFILE_UPDATED = 'USER_PROFILE_UPDATED';
const USER_PROFILE_REQUESTED = 'USER_PROFILE_REQUESTED';

let userProfileData = {
  fullName: 'Gini DevOps',
  email: 'ginidevops@gmail.com',
  address: {
    name: 'Gini DevOps',
    street: 'marsh addison drive texas 75287',
    city: 'United States',
    phone: ''
  },
  storeCredit: 0
};

// Fetch user profile from backend microservice
async function fetchUserProfile(userId) {
  try {
    const response = await fetch(`http://localhost:4600/api/v1/user/${userId}`);
    if (!response.ok) throw new Error('Failed to fetch user profile');
    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Error fetching user profile:', error);
    return userProfileData; // fallback to default
  }
}

// Update the UI
function updateUserInterface(data) {
  document.getElementById('user-fullname').textContent = data.fullName || 'Gini DevOps';
  document.getElementById('user-email').textContent = data.email || 'ginidevops@gmail.com';

  const addressElement = document.getElementById('default-address');
  if (addressElement && data.address) {
    addressElement.innerHTML = `
      <p>${data.address.name || 'Gini DevOps'}</p>
      <p>${data.address.street || 'marsh addison drive texas 75287'}</p>
      <p>${data.address.city || 'United States'}</p>
    `;
  }

  const creditAmount = document.querySelector('.credit-amount');
  if (creditAmount) {
    creditAmount.textContent = `$ ${data.storeCredit || 0}`;
  }

  const userProfileHeader = document.querySelector('#user-profile span');
  if (userProfileHeader) {
    userProfileHeader.textContent = `Hi, ${data.fullName?.split(' ')[0] || 'Gini'}`;
  }
}

// Notify other microfrontends
function notifyProfileUpdated(data) {
  const event = new CustomEvent(USER_PROFILE_UPDATED, {
    detail: {
      userName: data.fullName?.split(' ')[0] || 'Gini',
      email: data.email
    }
  });
  window.dispatchEvent(event);
}

// Listen for profile requests
window.addEventListener(USER_PROFILE_REQUESTED, async (event) => {
  const userId = event.detail?.userId || 'current';
  const data = await fetchUserProfile(userId);
  updateUserInterface(data);
  notifyProfileUpdated(data);
});

// Initialize
async function initializeUserProfile() {
  const userId = new URLSearchParams(window.location.search).get('userId') || 'current';
  const data = await fetchUserProfile(userId);
  userProfileData = data;
  updateUserInterface(data);
}

document.addEventListener('DOMContentLoaded', () => {
  // Sidebar navigation active state
  document.querySelectorAll('.nav-item a').forEach(link => {
    link.addEventListener('click', (e) => {
      document.querySelectorAll('.nav-item').forEach(item => item.classList.remove('active'));
      link.parentElement.classList.add('active');
    });
  });
  initializeUserProfile();
});

// Expose API
window.UserProfileMicrofrontend = {
  loadUserProfile: async (userId) => {
    const data = await fetchUserProfile(userId);
    updateUserInterface(data);
    return data;
  },
  updateUserField: async (field, value) => {
    userProfileData[field] = value;
    updateUserInterface(userProfileData);
    notifyProfileUpdated(userProfileData);
    return userProfileData;
  }
};

window.addEventListener('HOME_NAVIGATE_TO_PROFILE', (event) => {
  // Implement specific navigation behavior if needed
});
