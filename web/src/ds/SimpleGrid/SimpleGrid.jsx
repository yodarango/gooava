import "./SimpleGrid.css";

export function SimpleGrid({
  children,
  className = "",
  columns,
  minColWidth = "25rem",
  maxColWidth = "1fr",
  equalHeight = true,
  ...restOfProps
}) {
  const gridStyle = {
    display: "grid",
    gridTemplateColumns: columns
      ? `repeat(${columns}, 1fr)`
      : `repeat(auto-fill, minmax(${minColWidth}, ${maxColWidth}))`,
  };

  const gridClass = `simple-grid-34jt d-flex align-items-center justify-content-center flex-wrap gap-4 ${
    equalHeight ? "equal-height" : ""
  } ${className}`;

  return (
    <div className={gridClass} style={gridStyle} {...restOfProps}>
      {children}
    </div>
  );
}
