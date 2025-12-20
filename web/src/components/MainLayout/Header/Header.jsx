import { useAppContext } from "../../../views/context/appContextProvider";
import { GlobalSettings } from "../../GlobalSettings/GlobalSettings";
import { Link, generatePath, useParams } from "react-router-dom";
import { ROUTE_HOME, ROUTE_SESSIONS_ID } from "@constants";
import keeweeLogo from "../../../../public/logo.png";
import React from "react";
import { If } from "@ds";

// styles
import "./Header.css";

export const Header = () => {
  const { toggleMenu } = useAppContext();
  const { sessionId = "" } = useParams();

  const sessionPath = generatePath(ROUTE_SESSIONS_ID, {
    sessionId,
  });

  const shouldGlobalMenuShowRouter = [sessionPath];

  return (
    <>
      <div className='app-header-56yl__spacer'></div>
      <header className='app-header-56yl'>
        <div className='app-header-56yl__container'>
          <Link to={ROUTE_HOME} className='app-header-56yl__brand'>
            <div className='app-header-56yl__logo'>
              <img src={keeweeLogo} alt='Keewee Logo' className='logo' />
            </div>
            <span className='app-header-56yl__title'>Keewee</span>
          </Link>
          {/* Navigation */}
          <nav className='app-header-56yl__nav'>
            <If
              condition={shouldGlobalMenuShowRouter.includes(location.pathname)}
            >
              <button className='bg-nu ps-3 pe-1 py-0 m-0' onClick={toggleMenu}>
                <ion-icon name='ellipsis-vertical-outline'></ion-icon>
              </button>
            </If>
            <If
              condition={
                !shouldGlobalMenuShowRouter.includes(location.pathname)
              }
            >
              <GlobalSettings />
            </If>
          </nav>
        </div>
      </header>
    </>
  );
};
