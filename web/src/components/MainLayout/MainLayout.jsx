import { useAppContext } from "../../views/context/appContextProvider";
import { Outlet, useNavigate, useLocation } from "react-router-dom";
import { Header } from "./Header/Header";
import { Footer } from "./Footer/Footer";
import {
  ROUTE_AUTH_VERIFY,
  ROUTE_SESSIONS,
  ROUTE_HOME,
  ROUTE_AUTH,
} from "@constants";

// styles
import "./MainLayout.css";
import { useEffect } from "react";

export const MainLayout = () => {
  const { state } = useAppContext();
  const navigate = useNavigate();
  const location = useLocation();

  function redirectUser() {
    // If the user is logged in and this account needs to be verified before proceeding
    if (
      location.pathname === ROUTE_HOME &&
      state.isAuthenticated &&
      state.isPending
    ) {
      return navigate(ROUTE_AUTH_VERIFY);
    }

    // if the user is tryng to study but it is not logged in
    if (
      location.pathname.startsWith(ROUTE_SESSIONS) &&
      !state.isAuthenticated
    ) {
      return navigate(ROUTE_AUTH);
    }

    // if the user is logged in but is trying to hit th auth page
    if (location.pathname === ROUTE_AUTH_VERIFY && state.isActive) {
      return navigate(ROUTE_SESSIONS);
    }

    // if the user is NOT logged in but is trying to hit the email verification page
    if (location.pathname === ROUTE_AUTH_VERIFY && !state.isAuthenticated) {
      return navigate(ROUTE_AUTH);
    }

    if (location.pathname === ROUTE_AUTH && state.isAuthenticated) {
      return navigate(ROUTE_SESSIONS);
    }
  }

  useEffect(() => {
    if (state.isLoading) return;

    redirectUser();
  }, [location, state.isAuthenticated, state.isPending, state.isActive]);

  return (
    <div className='main-layout-56yl'>
      <Header />
      <main className='main-layout-56yl__content'>
        <Outlet />
      </main>
      <Footer />
    </div>
  );
};
