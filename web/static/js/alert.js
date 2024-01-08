/**
 * Returns a dismissable bootstrap alert. The content of the alert is message and it
 * is styled based on the type.
 * 
 * @param message The message to be displayed.
 * @param type The bootstrap alert style.
 * @returns A dissmisable bootstrap alert.
 */
const createAlert = (message, type) => {
    const wrapper = document.createElement('div');
    wrapper.innerHTML = [
        `<div class="alert alert-${type} alert-dismissible" role="alert">`,
        `   <div>${message}</div>`,
        '   <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>',
        '</div>'
    ].join('');

    return wrapper;
}

// The timeout for the active alert. Each time an alert is created, it will share the same underlying
// timeout. When another alert is displayed, if current timeout has not timed out, alertTimeout will be
// cleared and reset to 5 seconds.
let alertTimeout;

/**
 * Creates a bootstrap alert styled as type, with the content set to message.
 * This bootstrap alert will then be inserted into dest. The alert will remain open for
 * 5 seconds and then cleared, if it has not already been dismissed. Writing an alert while
 * another alert is open, will result in it being overwritten and the timeout reset.
 *
 * @param dest The destination element, this should be a div placeholder.
 * @param message The alert message.
 * @param type The alert type, typically danger or success.
 */
export function writeAlert(dest, message, type) {
    if (alertTimeout) {
        clearTimeout(alertTimeout);
    }

    dest.innerHTML = createAlert(message, type).innerHTML;

    alertTimeout = setTimeout(() => {
        dest.innerHTML = "";
    }, 5000);
}

/**
 * Clears the div placeholder that holds the bootstrap alert.
 *
 * @param dest The div placeholder that holds the bootstrap alert.
 */
export function clearAlert (dest) {
    dest.innerHTML = "";
}

// The alert placeholder div in the base layout template. This is the central place to display alerts.
const alertPlaceholder = document.getElementById("alertPlaceholder");

/**
 * Creates and writes a bootstrap success alert with the content set to message. The alert is written
 * into the alertPlaceholder div in the base layout template.
 * 
 * @param {string} message The message to be displayed in a bootstrap success alert div.
 */
export function writeFlashMessage(message) {
    writeAlert(alertPlaceholder, message, "success");
}

/**
 * Creates and writes a bootstrap danger alert with the content set to message. The alert is written
 * into the alertPlaceholder div in the base layout template.
 * 
 * @param {string} message The message to be displayed in a bootstrap danger alert div.
 */
export function writeFlashError(error) {
    writeAlert(alertPlaceholder, error, "danger");
}