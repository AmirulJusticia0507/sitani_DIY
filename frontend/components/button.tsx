export function Button({
  children,
  className = "",
  variant = "primary",
  disabled = false,
  onClick,
  type = "button",
  size = "md",
  title,
}: {
  children: React.ReactNode;
  className?: string;
  variant?: "primary" | "outline" | "secondary";
  disabled?: boolean;
  onClick?: () => void;
  type?: "button" | "submit" | "reset";
  size?: "sm" | "md";
  title?: string;
}) {
  const base = "btn";
  const variants: Record<string, string> = {
    primary: "btn-primary",
    outline: "btn-outline",
    secondary: "btn-secondary",
  };
  const sizes: Record<string, string> = {
    sm: "btn-sm",
    md: "",
  };

  return (
    <button
      type={type}
      className={[base, variants[variant], sizes[size], className].filter(Boolean).join(" ")}
      disabled={disabled}
      onClick={onClick}
      title={title}
    >
      {children}
    </button>
  );
}
