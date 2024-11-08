function copyTimeToClipboard() {
    const passwordText = document.getElementById("utc-timestamp").innerText;
    navigator.clipboard.writeText(passwordText).then(() => {
        // Show the tooltip
        const tooltip = document.getElementById("copy-tooltip");
        tooltip.style.visibility = "visible";

        // Hide the tooltip after 1 second
        setTimeout(() => {
            tooltip.style.visibility = "hidden";
        }, 1000);
    }).catch(err => {
        console.error('Failed to copy: ', err);
    });
    navigator.clipboard.writeText(passwordText);
}

function copyPwToClipboard() {
    const passwordText = document.getElementById("password-display").innerText;
    navigator.clipboard.writeText(passwordText).then(() => {
        // Show the tooltip
        const tooltip = document.getElementById("copy-tooltip");
        tooltip.style.visibility = "visible";

        // Hide the tooltip after 1 second
        setTimeout(() => {
            tooltip.style.visibility = "hidden";
        }, 1000);
    }).catch(err => {
        console.error('Failed to copy: ', err);
    });
}

