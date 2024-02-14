const ErrorManager = {
    ShowError: (message) => {
        const content = document.getElementById("errorMessage");
        content.style.display = "block";
        content.innerText = message;
    },

    ClearError: () => {
        const content = document.getElementById("errorMessage");
        content.style.display = "none";
    }
}
