import React, { useState } from "react";
import "./Tabs.css";

export const Tabs = ({ children }) => {
  const tabTitles = React.Children.toArray(children).filter(
    (child) => child.type.displayName === "TabItem",
  );
  const tabContents = React.Children.toArray(children).filter(
    (child) => child.type.displayName === "TabContent",
  );
  const [activeIndex, setActiveIndex] = useState(0);

  return (
    <div className='tabs-container'>
      <div className='tabs-container__titles'>
        {tabTitles.map((tab, idx) => (
          <button
            key={idx}
            className={`tabs-container__tab${
              activeIndex === idx ? " active" : ""
            }`}
            onClick={() => setActiveIndex(idx)}
          >
            {tab.props.children}
          </button>
        ))}
      </div>
      <div className='tabs-container__content'>{tabContents[activeIndex]}</div>
    </div>
  );
};

export const TabItem = ({ children }) => children;
TabItem.displayName = "TabItem";

export const TabContent = ({ children }) => <div>{children}</div>;
TabContent.displayName = "TabContent";
