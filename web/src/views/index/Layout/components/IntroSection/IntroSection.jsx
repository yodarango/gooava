import welcomeImage from "@images/statuses/welcome.webp";
import { Link } from "react-router-dom";
import { Button } from "@ds";
import { ROUTE_AUTH } from "@constants";

// styles
import "./IntroSection.css";

export const IntroSection = () => {
  return (
    <section className='intro-section-56yl'>
      {/* Hero Section */}
      <div className='intro-hero'>
        <div className='intro-hero__content'>
          <h1 className='intro-hero__title'>
            Welcome to <span className='color-beta'>Goilerplate</span>
          </h1>
          <div className='intro-hero__image'>
            <img
              src={welcomeImage}
              alt='Welcome to Goilerplate'
              className='welcome-image'
            />
          </div>
          <p className='intro-hero__subtitle mb-2'>
            A clean starting point for full-stack Go and React applications.
          </p>
        </div>
      </div>

      {/* CTA Section */}
      <div className='intro-cta text-center p-5 mb-6'>
        <h2 className='mb-2'>Ready to build?</h2>
        <p className='mb-4'>
          Authentication, database connectivity, and a modern UI are already
          wired up. Start adding your own features.
        </p>
        <div className='d-flex align-items-center justify-content-center gap-4'>
          <Link to={ROUTE_AUTH} className='w-100'>
            <Button secondary className='cta-button w-100'>
              Sign In or Create Account
            </Button>
          </Link>
        </div>
      </div>
    </section>
  );
};
