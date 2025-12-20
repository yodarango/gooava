import { WORD_GRADINGS_SECONDS, ROUTE_SESSIONS } from "@constants";
import welcomeImage from "@images/statuses/welcome.webp";
import React, { useState, useEffect } from "react";
import { Link } from "react-router-dom";
import { Button } from "@ds";

// styles
import "./IntroSection.css";

export const IntroSection = () => {
  const [highlightedLevel, setHighlightedLevel] = useState(1);

  useEffect(() => {
    const timer = setInterval(() => {
      setHighlightedLevel((prev) => (prev % 22) + 1);
    }, 1000);

    return () => clearInterval(timer);
  }, []);

  // SRS Level colors (22 levels)
  const srsLevels = WORD_GRADINGS_SECONDS;

  return (
    <section className='intro-section-56yl'>
      {/* Hero Section */}
      <div className='intro-hero'>
        <div className='intro-hero__content'>
          <h1 className='intro-hero__title'>
            Welcome to <span className='color-beta'>Keewee</span>
          </h1>
          <div className='intro-hero__image'>
            <img
              src={welcomeImage}
              alt='Welcome to Keewee'
              className='welcome-image'
            />
          </div>
          <p className='intro-hero__subtitle mb-2'>
            Master languages through intelligent spaced repetition and
            traditional flashcards
          </p>
        </div>
      </div>

      {/* CTA Section */}
      <div className='intro-cta text-center p-5 mb-6'>
        <h2 className='mb-2'>Pick your poison</h2>
        <p className='mb-4'>
          Is patience your middle name? You probably will enjoy the flashcards
          feature. Do you eat discipline for breakfast? Go with the SRS
        </p>
        <div className='d-flex align-items-center justify-content-center gap-4'>
          <Link to={ROUTE_SESSIONS} className='w-100'>
            <Button secondary className='cta-button w-100'>
              <ion-icon name='library-outline'></ion-icon>
              Try SRS
            </Button>
          </Link>
        </div>
      </div>

      {/* App Overview */}
      <div className='intro-overview'>
        <div className='intro-card'>
          <h2>🥝 What is Keewee?</h2>
          <p>
            Nothing fancy. Just another <b>Spaced Repetition System (SRS)</b> to
            scartch your bilingual, trilingual, or polyglot itch. What makes{" "}
            <b>Keewee</b> different? Proably the most different factor is its
            pre-made card sets for the most popular languages, intellengtly
            arranged in popularity order and cross compatible with each other.
            Long gone are the days of you having to find a deck from Arabic to
            Japanese or from Dutch to Hindi.
          </p>
        </div>
      </div>

      {/* SRS Explanation */}
      <div className='intro-srs'>
        <div className='intro-card'>
          <h2>🧠 How SRS Works</h2>
          <p>
            The Spaced Repetition System is a learning technique that increases
            intervals between reviews of previously learned material. When you
            answer correctly, the word moves to the next level with a longer
            interval. If you struggle, it drops back to reinforce learning.
          </p>
          <h5 className='mb-2'>Benefits</h5>
          <div className='srs-benefits'>
            <div className='benefit-item'>
              <span className='benefit-icon'>⏰</span>
              <span>Optimized review timing</span>
            </div>
            <div className='benefit-item'>
              <span className='benefit-icon'>🎯</span>
              <span>Focus on difficult words</span>
            </div>
            <div className='benefit-item'>
              <span className='benefit-icon'>🧠</span>
              <span>Long-term retention</span>
            </div>
          </div>
        </div>
      </div>

      {/* 22 Levels System */}
      <div className='intro-levels mb-6'>
        <div className='success-card'>
          <h2>🏆 The 22-Level Journey</h2>
          <p>
            Every word in Keewee progresses through <strong>22 levels </strong>
            before graduation where each level represents mastery milestones
            with increasing intervals between reviews. It might look
            intimidating, but don't you worry, the journey is much easier than
            it looks!
          </p>
          <div className='levels-grid'>
            {srsLevels.map((level) => (
              <div
                key={level.value}
                className={`level-badge ${
                  highlightedLevel === level.value ? "highlighted" : ""
                }`}
                style={{
                  backgroundColor: level.color,
                }}
              >
                <span className='level-number fs-5 mb-1'>{level.value}</span>
                <span className='level-name fs-6'>{level.label}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Traditional Flashcards */}
      <div className='intro-flashcards'>
        <div className='intro-card'>
          <h2>📚 Traditional Flashcards</h2>
          <p>
            If you are a lover of simplicity, you might enjoy taking a look at
            the flashcards feature instead, which allos you to choose from a ton
            pre-made card sets or upload your own.
          </p>
          <ul className='flashcard-features'>
            <li>⏰ Study at your own pace</li>
            <li>🎣 Choose cards by language</li>
            <li>⬆️ Upload your own cards</li>
            <li>🎯 Focus on specific topics</li>
          </ul>
        </div>
      </div>
    </section>
  );
};
