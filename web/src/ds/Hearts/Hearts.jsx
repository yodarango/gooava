import { useEffect, useRef } from "react";
import "./Hearts.css";

export const Hearts = ({
  autoStart = true,
  enableSound = true,
  enableClickHearts = true,
  enableAutoHearts = true,
  heartColors = ["#456", "#890", "#634", "#299", "tomato", "#fb3"],
  className = "",
}) => {
  const containerRef = useRef(null);
  const audioContextRef = useRef(null);
  const intervalRef = useRef(null);

  function createSound(size, fr, delay, type, vol) {
    if (!enableSound || !audioContextRef.current) return;

    for (let i = 0; i < size; i++) {
      let osc = audioContextRef.current.createOscillator(),
        gain = audioContextRef.current.createGain();

      setTimeout(function () {
        osc.frequency.value = fr * i;
        gain.gain.value = vol;
        osc.type = type;
        osc.connect(gain);
        gain.connect(audioContextRef.current.destination);
        osc.start();
        setTimeout(function () {
          let gVal = gain.gain.value,
            smooth;
          function reduceGain() {
            gVal -= 0.02;
            if (gVal > 0) {
              smooth = requestAnimationFrame(reduceGain);
            } else {
              osc.stop();
              cancelAnimationFrame(smooth);
            }
            gain.gain.value = gVal / 7;
          }
          reduceGain();
        }, delay);
      }, i * delay);
    }
  }

  function getRandomColor() {
    const idx = Math.floor(heartColors.length * Math.random());
    return heartColors[idx];
  }

  function animateIt(el, dur, delay) {
    var animateEl = el.animate(
      [
        {
          opacity: 0,
          transform: "translate(-50%, -50%) scale(0)",
        },

        {
          opacity: 1,
          transform: "translate(-50%, -50%) scale(1)",
        },
        {
          opacity: 0,
          transform: "translate(-50%, -50%) scale(1.1)",
        },
      ],
      {
        duration: dur,
        easing: "ease-out",
        fill: "forwards",
        delay: delay || 0,
      }
    );

    return animateEl;
  }

  function createBubble() {
    if (!containerRef.current) return null;

    const ns = "http://www.w3.org/2000/svg";
    const bubble = document.createElement("div");
    const bubbleDummy = document.createElement("div");
    const heart = document.createElementNS(ns, "svg");
    const heartPath = document.createElementNS(ns, "path");

    heart.setAttribute("viewBox", "0 0 600 500");
    heartPath.setAttribute(
      "d",
      "M300,150 C500,10 600,300 300,400 C0,300 100,10 300,150"
    );
    bubble.classList.add("hearts-bubble");
    bubble.style.color = getRandomColor();
    bubbleDummy.classList.add("hearts-bubble-dummy");
    heart.classList.add("hearts-heart");

    heart.appendChild(heartPath);
    bubble.appendChild(bubbleDummy);
    bubble.appendChild(heart);

    containerRef.current.appendChild(bubble);

    return {
      setPosition: function (x, y) {
        bubble.style.left = x + "px";
        bubble.style.top = y + "px";
      },
      _animate: function () {
        const animateBubble = animateIt(bubbleDummy, 1200);
        const animateHeart = animateIt(heart, 2000);

        return {
          bubbleDur: 1200,
          heartDur: 2000,
        };
      },
      remove: function () {
        bubble.remove();
      },
    };
  }

  function handleDown(e) {
    if (!enableClickHearts) return;

    const rect = containerRef.current?.getBoundingClientRect();
    if (!rect) return;

    const _x = e.pageX - rect.left;
    const _y = e.pageY - rect.top;

    const bubble = createBubble();
    if (!bubble) return;

    bubble.setPosition(_x, _y);
    const animation = bubble._animate();
    const totalDelay = animation.bubbleDur + animation.heartDur;

    if (e.type && enableSound) {
      createSound(20, 5000, 1, "sawtooth", 1);
    }

    setTimeout(() => {
      bubble.remove();
    }, totalDelay);
  }

  function bubbleUp() {
    if (!enableAutoHearts || !containerRef.current) return;

    const rect = containerRef.current.getBoundingClientRect();
    const de = {
      pageX: Math.random() * rect.width + rect.left,
      pageY: Math.random() * rect.height + rect.top,
    };

    handleDown(de);
  }

  useEffect(() => {
    if (!autoStart) return;

    // Initialize audio context
    if (enableSound) {
      try {
        audioContextRef.current = new (window.AudioContext ||
          window.webkitAudioContext)();
      } catch (e) {
        console.warn("Audio context not supported");
      }
    }

    // Add click listeners
    const container = containerRef.current;
    if (container && enableClickHearts) {
      container.addEventListener("click", handleDown, false);
      container.addEventListener("touchstart", handleDown, false);
    }

    // Start auto hearts
    if (enableAutoHearts) {
      intervalRef.current = setInterval(bubbleUp, 200);
    }

    // Cleanup
    return () => {
      if (container && enableClickHearts) {
        container.removeEventListener("click", handleDown, false);
        container.removeEventListener("touchstart", handleDown, false);
      }
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
      if (audioContextRef.current) {
        audioContextRef.current.close();
      }
    };
  }, [autoStart, enableSound, enableClickHearts, enableAutoHearts]);

  return <div ref={containerRef} className={`hearts-container ${className}`} />;
};
