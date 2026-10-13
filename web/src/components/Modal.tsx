import type { ReactNode } from "react";

interface ModalProps {
  title: string;
  open: boolean;
  onClose: () => void;
  children: ReactNode;
  actions?: ReactNode;
}

export function Modal({ title, open, onClose, children, actions }: ModalProps) {
  if (!open) return null;

  return (
    <div className="popup-overlay" onClick={onClose}>
      <div className="popup" onClick={(e) => e.stopPropagation()}>
        <h2>{title}</h2>
        {children}
        {actions && <div className="popup-actions">{actions}</div>}
      </div>
    </div>
  );
}

interface PopupButtonProps {
  children: ReactNode;
  onClick?: () => void;
  variant?: "primary" | "secondary" | "destructive";
  disabled?: boolean;
}

export function PopupButton({
  children,
  onClick,
  variant = "secondary",
  disabled,
}: PopupButtonProps) {
  const className =
    variant === "destructive"
      ? "destructive"
      : variant === "secondary"
        ? "secondary"
        : "";
  return (
    <button type="button" className={className} onClick={onClick} disabled={disabled}>
      {children}
    </button>
  );
}
