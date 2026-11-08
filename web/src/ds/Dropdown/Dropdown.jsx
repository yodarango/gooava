import { PopupMenu, PopupMenuItem } from "../PopupMenu/PopupMenu";
import { ChevronDown } from "lucide"; // data, not a component
import { MorphIcon } from "morphicons/react";

// styles
import "./Dropdown.css";

/*********************************************************************************************************
 * Design-system dropdown: a pill trigger showing the current value that opens a popup
 * menu of options. options = [{ value, label }]; onChange receives the picked value.
 * ******************************************************************************************************
 */
export const Dropdown = (props) => {
  const {
    value,
    options,
    onChange,
    ariaLabel = "Select an option",
    className = "",
    secondary = false, // white pill w/ shadow — matches the Button secondary variant
  } = props;

  const current = options.find((option) => option.value === value);

  return (
    <PopupMenu
      closeOnSelect
      trigger={
        <button
          className={`ds-dropdown ${secondary ? "secondary" : ""} ${className}`.trim()
          }
          type='button'
          aria-label={ariaLabel}
          title={ariaLabel}
        >
          <span className='ds-dropdown__value'>
            {current ? current.label : "—"}
          </span>
          <MorphIcon icon={ChevronDown} label='Open options' size={16} />
        </button>
      }
    >
      {options.map((option) => (
        <PopupMenuItem key={option.value} onClick={() => onChange(option.value)}>
          {option.label}
        </PopupMenuItem>
      ))}
    </PopupMenu>
  );
};
