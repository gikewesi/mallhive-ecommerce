const BASE_URL = "http://127.0.0.1:5000";

// Register a new user
async function registerUser(email, password, name) {
  const response = await fetch(`${BASE_URL}/register`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password, name }),
  });
  const data = await response.json();
  alert(data.message);
}

// Login a user
async function loginUser(email, password) {
  const response = await fetch(`${BASE_URL}/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
  });
  const data = await response.json();
  
  if (response.ok) {
      localStorage.setItem("accessToken", data.access_token);
      alert("Login successful!");
  } else {
      alert(data.message);
  }
}

// Fetch profile data
async function fetchProfile() {
  const token = localStorage.getItem("accessToken");
  
  const response = await fetch(`${BASE_URL}/profile`, {
      method: "GET",
      headers: { Authorization: `Bearer ${token}` },
  });
  
  const data = await response.json();
  
  if (response.ok) {
      console.log(data);
      alert(`Name: ${data.name}, Email: ${data.email}`);
  } else {
      alert(data.message);
  }
}
