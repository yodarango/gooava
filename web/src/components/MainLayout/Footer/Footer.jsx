// styles
import "./Footer.css";

export const Footer = () => {
  const currentYear = new Date().getFullYear();

  return (
    <footer className='app-footer-56yl'>
      <div className='app-footer-56yl__container'>
        <p className='app-footer-56yl__text'>
          © {currentYear} Goilerplate. Built as a starting point for your next
          app.
        </p>
      </div>
    </footer>
  );
};
