const InfoManager = {
    ShowInfo: (message) => {
        const content = document.getElementById("infoMessage");
        content.style.display = "block";
        content.innerText = message;
    },

    ClearInfo: () => {
        const content = document.getElementById("infoMessage");
        content.style.display = "none";
    }
}
