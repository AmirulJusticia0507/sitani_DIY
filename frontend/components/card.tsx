export function Card({
  children,
  className = "",
  hover = false,
}: {
  children: React.ReactNode;
  className?: string;
  hover?: boolean;
}) {
  return (
    <section className={["card", hover ? "card-hover" : "", className].filter(Boolean).join(" ")}>
      {children}
    </section>
  );
}
