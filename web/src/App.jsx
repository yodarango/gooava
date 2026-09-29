import {
  IndexView,
  AuthView,
  AuthVerifyView,
  FinancesView,
  CalendarView,
} from "@views";
import { MainLayout, ProtectedRoute, AuthRoute } from "@components";
import {
  createRoutesFromElements,
  createBrowserRouter,
  RouterProvider,
  Route,
  Outlet,
} from "react-router-dom";
import {
  ROUTE_AUTH_VERIFY,
  ROUTE_FINANCES,
  ROUTE_CALENDAR,
  ROUTE_HOME,
  ROUTE_AUTH,
} from "@constants";

// global styles
import "@assets/tokens.css";
import "./App.css";

const router = createBrowserRouter(
  createRoutesFromElements(
    <Route element={<Outlet />}>
      {/* Everything behind auth: app shell + pages */}
      <Route element={<ProtectedRoute />}>
        <Route element={<MainLayout />} errorElement={<></>}>
          <Route path={ROUTE_HOME} element={<IndexView />} />
          <Route path={ROUTE_FINANCES} element={<FinancesView />} />
          <Route path={ROUTE_CALENDAR} element={<CalendarView />} />
        </Route>
      </Route>

      {/* Auth pages handle their own redirects based on user status */}
      <Route element={<AuthRoute />}>
        <Route path={ROUTE_AUTH} element={<AuthView />} />
        <Route path={ROUTE_AUTH_VERIFY} element={<AuthVerifyView />} />
      </Route>
    </Route>,
  ),
);

export default function App() {
  return <RouterProvider router={router} />;
}
