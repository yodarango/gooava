import "./SectionalButton.css";

export const SectionalButtom = (props) => {
  const {
    className,
    secondary,
    gradient,
    children,
    primary,
    label,
    icon,
    value,
    ...rest
  } = props;

  const variantClass = primary
    ? "primary"
    : secondary
    ? "secondary"
    : gradient
    ? "gradient"
    : "";

  return (
    <button
      className={`sectional-button-00it ${variantClass} ${className}`}
      {...rest}
    >
      <span>{label}</span>
      <span className='d-inline-flex align-items-center color-alpha justify-content-end gap-2'>
        <span>{value}</span>
        <ion-icon name={icon} />
      </span>
    </button>
  );
};
