import welcomeImage from "@images/statuses/welcome.webp";

// styles
import "./IntroSection.css";

export const IntroSection = () => {
  return (
    <section className='intro-section-56yl'>
      {/* Hero Section */}
      <div className='intro-hero'>
        <div className='intro-hero__content'>
          <h1 className='intro-hero__title'>
            Welcome to <span className='color-beta'>gooava</span>
          </h1>
          <div className='intro-hero__image'>
            <img
              src={welcomeImage}
              alt='Welcome to gooava'
              className='welcome-image'
            />
          </div>
          <p className='intro-hero__subtitle mb-2'>
            Your personal home management hub — finances, meals, and everything
            in between.
          </p>
        </div>
      </div>
    </section>
  );
};
