type BadgeVariant = "success" | "warning" | "danger" | "info";

export function Badge({
  children,
  variant = "success",
  className = "",
}: {
  children: React.ReactNode;
  variant?: BadgeVariant;
  className?: string;
}) {
  const variants: Record<string, string> = {
    success: "badge-success",
    warning: "badge-warning",
    danger: "badge-danger",
    info: "badge-info",
  };

  return (
    <span className={["badge", variants[variant], className].filter(Boolean).join(" ")}>
      {children}
    </span>
  );
}
