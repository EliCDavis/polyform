const infoMessageId = "infoMessage";

export const InfoManager = {
    ShowInfo: (message: string) => {
        const content = document.getElementById(infoMessageId);
        content.style.display = "block";
        content.innerText = message;
    },

    ClearInfo: () => {
        const content = document.getElementById(infoMessageId);
        content.style.display = "none";
    }
}
