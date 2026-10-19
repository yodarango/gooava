import { useAppContext } from "../../../views/context/appContextProvider";
import { ROUTE_HOME, ROUTE_ME } from "@constants";
import { Link } from "react-router-dom";
import { Thumbnail } from "@ds";
import { avatars } from "@images";
import gooavaLogo from "../../../../public/logo.png";

// styles
import "./Header.css";

export const Header = () => {
  const { state } = useAppContext();
  const { user } = state;

  const avatar = avatars.find((a) => a.pathName === user?.avatar) || null;

  return (
    <>
      <div className='app-header-56yl__spacer'></div>
      <header className='app-header-56yl'>
        <div className='app-header-56yl__container'>
          <Link to={ROUTE_HOME} className='app-header-56yl__brand'>
            <div className='app-header-56yl__logo'>
              <img src={gooavaLogo} alt='Gooava Logo' className='logo' />
            </div>
          </Link>

          {/* profile shortcut — avatar + username, links to the /me page */}
          <Link
            to={ROUTE_ME}
            className='app-header-56yl__avatar-container'
            aria-label='My profile'
          >
            {avatar ? (
              <Thumbnail
                src={avatar.image}
                alt={`${user.first_name || "User"}'s avatar`}
                className='app-header-56yl__avatar'
                width='40px'
              />
            ) : (
              <span className='app-header-56yl__avatar-fallback'>
                {(user?.first_name?.[0] || "?").toUpperCase()}
              </span>
            )}
            <span className='app-header-56yl__username'>{user?.username}</span>
          </Link>
        </div>
      </header>
    </>
  );
};
