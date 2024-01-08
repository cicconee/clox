import { writeAlert } from "./alert.js";

// The registration form that is submitted when registering a user account.
const registerForm = document.getElementById("registerForm");

// The username textfield in the registration form.
const username = document.getElementById("username");

// The div that is the alert placeholder.
const alertPlaceholder = document.getElementById("alertPlaceholder");

// Set the registration form submit event handler. The registration form is posted to the server.
registerForm.addEventListener("submit", function(e) {
    e.preventDefault();

    const formData = new FormData(this);

    fetch(this.action, {
        method: this.method,
        headers: {
            "X-Requested-With": "FetchAPI"
        },
        body: formData,
        credentials: "include"
    })
    .then(resp => {
        if (!resp.ok) {
            // Clear username input value on failed request.
            username.value = "";

            return resp.json().then(errData => {
                throw errData;
            })
        }
        
        if (resp.redirected) {
            window.location.href = resp.url;
        }
    })
    .catch(errData => {
        writeAlert(alertPlaceholder, errData.error, "danger");
    })
});