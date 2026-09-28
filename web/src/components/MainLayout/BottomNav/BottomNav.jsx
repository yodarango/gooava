import { useAppContext } from "../../../views/context/appContextProvider";
import { UtensilsCrossed, CookingPot, House, DoorOpen, Receipt, ReceiptText } from "lucide";
import { useLocation, useNavigate } from "react-router-dom";
import { MorphIcon } from "morphicons/react";
import { ROUTE_HOME } from "@constants";
import { useState } from "react";

// styles
import "./BottomNav.css";

export const BottomNav = () => {
  const { showToast } = useAppContext();
  const location = useLocation();
  const navigate = useNavigate();

  const [hoveredItem, setHoveredItem] = useState(null);

  const isHomeActive = location.pathname === ROUTE_HOME;

  const comingSoon = (feature) => () =>
    showToast({ message: `${feature} are coming soon!`, type: "info" });

  return (
    <nav className='bottom-nav-9x2k' aria-label='Main navigation'>
      <div className='bottom-nav-9x2k__pill'>
        <button
          onMouseEnter={() => setHoveredItem("meals")}
          onMouseLeave={() => setHoveredItem(null)}
          className='bottom-nav-9x2k__item'
          onClick={comingSoon("Meals")}
          aria-label='Meals'
          type='button'
        >
          <MorphIcon
            icon={hoveredItem === "meals" ? CookingPot : UtensilsCrossed}
            spring='snappy'
            label='Meals'
            size={24}
          />
        </button>

        <button
          onMouseEnter={() => setHoveredItem("home")}
          onMouseLeave={() => setHoveredItem(null)}
          className={`bottom-nav-9x2k__item bottom-nav-9x2k__item--home ${isHomeActive ? "active" : ""}`}
          onClick={() => navigate(ROUTE_HOME)}
          aria-label='Home'
          type='button'
        >
          <MorphIcon
            icon={
              hoveredItem === "home" || isHomeActive ? DoorOpen : House
            }
            spring='snappy'
            label='Home'
            size={26}
          />
        </button>

        <button
          onMouseEnter={() => setHoveredItem("bills")}
          onMouseLeave={() => setHoveredItem(null)}
          className='bottom-nav-9x2k__item'
          onClick={comingSoon("Bills")}
          aria-label='Bills'
          type='button'
        >
          <MorphIcon
            icon={hoveredItem === "bills" ? ReceiptText : Receipt}
            spring='snappy'
            label='Bills'
            size={24}
          />
        </button>
      </div>
    </nav>
  );
};
