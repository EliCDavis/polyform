import { useMessageStore } from "@/stores/messageStore";

export function MessageOverlay() {
  const errors = useMessageStore((s) => s.errors);
  const info = useMessageStore((s) => s.info);

  return (
    <div id="messageContainer">
      {Object.entries(errors).map(([key, message]) => (
        <div key={key} className="errorMessage">
          {message}
        </div>
      ))}
      {info !== null && <div id="infoMessage">{info}</div>}
    </div>
  );
}
