import { BottomNav } from "./BottomNav/BottomNav";
import { Header } from "./Header/Header";
import { Outlet } from "react-router-dom";

// styles
import "./MainLayout.css";

export const MainLayout = () => {
  return (
    <div className='main-layout-56yl'>
      <Header />
      <main className='main-layout-56yl__content'>
        <Outlet />
      </main>
      <BottomNav />
    </div>
  );
};
