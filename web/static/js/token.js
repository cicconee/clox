import { writeAlert, clearAlert, writeFlashError } from "./alert.js";

// The first dropdown item in the dropdown menu of the token form.
const defaultExpireOption = document.querySelector("#dropdownExpireList .dropdown-item");

// The expiration dropdown menu button in the token form.
const dropdownExpireButton = document.getElementById("dropdownExpireButton");

// The hidden input in the token form that holds the selected expiration time.
const selectedExpireValue = document.getElementById("selectedExpireValue");

// Set the default expire option.
setDefaultExpiration();

/**
 * Sets the default expiration value in the token form.
 */
function setDefaultExpiration() {
    setExpiration(defaultExpireOption.textContent, defaultExpireOption.getAttribute("data-value"));
}

// All of the dropdown items available in the dropdown menu of the token form.
const expireOptions = document.querySelectorAll("#dropdownExpireList .dropdown-item");

// Set the click event listener for every expiration option in the drop down menu.
expireOptions.forEach(item => {
    item.addEventListener("click", function() {
        setExpiration(this.textContent, this.getAttribute("data-value"));
    })
})

/**
 * Sets the expiration value in the token form.
 * 
 * @param text The displayed text of the selected expires option.
 * @param value The underlying value of the selected expires option that will be posted.
 */
function setExpiration(text, value) {
    dropdownExpireButton.textContent = text;
    selectedExpireValue.value = value;
}

// The token form that is submitted when generating a new token.
const tokenForm = document.getElementById("tokenForm");

// Set the token forms submit event listener. The token form is posted to the server.
tokenForm.addEventListener("submit", function(e) {
    e.preventDefault();

    const formData = new FormData(this);

    for (let [key, value] of formData.entries()) {
        console.log(key, value);
    }

    fetch("/tokens", {
        method: "POST",
        body: formData,
    })
    .then(resp => {
        if (!resp.ok) {
            return resp.json().then(errData => {
                throw errData;
            })
        }

        return resp.json();
    })
    .then(data => {
        console.log(data);

        const tokenFormModal = bootstrap.Modal.getInstance(document.getElementById("tokenFormModal"))
        tokenFormModal.hide();
        reset();    

        const tokenKeyModal = new bootstrap.Modal(document.getElementById('tokenValueModal'), {
            keyboard: false
        });
        tokenKeyModal.show();

        document.getElementById("tokenValue").innerText = data.token;

        // Get the table body from tokenTable.
        let tokenTableBody = document.getElementById("tokenTable").getElementsByTagName("tbody")[0];

        // Insert a new row at the end of the table body.
        let row = tokenTableBody.insertRow(-1);
        row.id = data["token_id"];

        // Insert a new cell (<th>) at the 0th index of the row for the token name.
        let nameCell = document.createElement("th");
        nameCell.scope = "row";
        nameCell.innerHTML = data["token_name"];
        row.appendChild(nameCell);

        // Insert remaining cells to the end of the row.
        let createdAtCell = row.insertCell(-1);
        let lastUsedCell = row.insertCell(-1);
        let expiresCell = row.insertCell(-1);
        let buttonCell = row.insertCell(-1);

        // Set the content of the remaining cells.
        createdAtCell.innerHTML = formatTime(data["created_at"]);
        lastUsedCell.innerHTML = formatTime(data["last_used"]);
        expiresCell.innerHTML = formatTime(data["expires_at"]);
        buttonCell.innerHTML = `
            <div class="dropdown" data-bs-toggle="dropdown">
                <button class="btn p-0"><i class="bi bi-three-dots h3"></i></button>
                <ul class="dropdown-menu">
                    <li><button class="dropdown-item" data-bs-toggle="modal" data-bs-target="#deleteTokenModal">Delete</button></li>
                </ul>
            </div>`;
    })
    .catch(errData => {
        console.log(errData);
        writeAlert(tokenModalAlertPlaceholder, errData.error, "danger");
    })
});

// Set the generate token button's click event listener to submit the token form.
document.getElementById("generateTokenButton").addEventListener("click", function() {
    tokenForm.dispatchEvent(new Event("submit", {
        bubbles: true,
        cancelable: true,
    }));
});

// The div that is the alert placeholder in the token form modal.
const tokenModalAlertPlaceholder = document.getElementById("tokenModalAlertPlaceholder");

// Set the cancel button's click event listener to reset the token form and modal.
document.getElementById("cancelTokenButton").addEventListener("click", reset);

// Set the header dismiss button's click event listener to reset the token form and modal.
document.getElementById("headerCancelTokenButton").addEventListener("click", reset);

/**
 * Resets the token form and modal to its default values.
 */
function reset() {
    clearAlert(tokenModalAlertPlaceholder);
    tokenForm.reset();
    setDefaultExpiration();
}

// Format all the times and set to local time zone.
document.querySelectorAll(".time").forEach(function(e) {
    e.textContent = formatTime(e.textContent);
})

/**
 * Converts a date and time string into a local date and time string. If timeString is a zero
 * date, "Never" is returned. This should only be the case for the "Last Used" column in the
 * token table.
 * 
 * @param {string} timeString A date and time string to be converted to a local date and time.
 * @returns The date and time string converted to the local time zone.
 */
function formatTime(timeString) {
    const date = new Date(timeString);

    // If the time is not set, default to "Never". This will only be the case for
    // the Last Used time. Otherwise, convert the time to the local time zone.
    if (date.getFullYear() === 0) {
        return "Never";
    } else {
        return date.toLocaleString();
    }
}

// The token ID that the action should be executed for. When a action is chosen from
// the action dropdown menu (three dots), it will be executed on behalf of this token ID.
let tokenID;

// Set the token tables click event listener to set the tokenID when the action dropdown 
// menu is clicked. 
document.getElementById("tokenTable").addEventListener("click", function(e) {
    // Set tokenID to the ID of <tr> when the action dropdown menu (three dots) is
    // clicked.
    if (e.target && e.target.classList.contains("bi-three-dots")) {
        const row = e.target.closest("tr");
        if (row && row.id) {
            tokenID = row.id;
        }
    }
});

// Set the confirmDeleteTokenButton's click event listener to delete the token from the table
// and close the bootstrap modal.
document.getElementById("confirmDeleteTokenButton").addEventListener("click", function() {
    const deleteTokenModal = document.getElementById("deleteTokenModal");
    bootstrap.Modal.getInstance(deleteTokenModal).hide();
    deleteToken(tokenID);
});

/**
 * Delete the token with the specified id. It is expected that the token id is also the id 
 * of the <tr> in the token table.
 * 
 * @param {string} id The id of the token to be deleted.
 */
function deleteToken(id) {
    const tokenRow = document.getElementById(id);
    const tokenResourceURL = getTokenResourceURL(id);
    
    if (tokenRow) {
        fetch(tokenResourceURL, {
            method: "DELETE",
            headers: {
                "X-Requested-With": "FetchAPI"
            },
            credentials: "include"
        })
        .then(resp => {
            if (!resp.ok) {
                return resp.json().then(errData => {
                    throw errData;
                })
            }
            
            tokenRow.remove();
        })
        .catch(errData => {
            writeFlashError(errData.error);
        })
    }
}

/**
 * Returns the token resource url for the provided token id. The token resource url format is expected
 * to be stored in a script tag with an id 'tokenResourceURL' and an attribute 'data-url' that holds the
 * format of the token resource url. 
 * 
 * <script id="tokenResourceURL" data-url="/tokens/{id}"></script>
 * 
 * @param {string} id The id of the token to get the resource url for.
 * @returns The token resource url.
 */
function getTokenResourceURL(id) {
    const tokenResourceURL = document.getElementById("tokenResourceURL").getAttribute("data-url");
    const decodedTokenResourceURL = decodeURIComponent(tokenResourceURL);
    return decodedTokenResourceURL.replace("{id}", id);
}