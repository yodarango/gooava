const languageFlagMap = [
  { names: ["english", "eng", "en", "English", "ENGLISH"], flag: "🇬🇧" },
  {
    names: ["spanish", "spa", "es", "español", "Español", "SPANISH"],
    flag: "🇪🇸",
  },
  {
    names: ["french", "fra", "fr", "français", "Français", "FRENCH"],
    flag: "🇫🇷",
  },
  {
    names: ["german", "deu", "de", "deutsch", "Deutsch", "GERMAN"],
    flag: "🇩🇪",
  },
  {
    names: ["italian", "ita", "it", "italiano", "Italiano", "ITALIAN"],
    flag: "🇮🇹",
  },
  {
    names: ["portuguese", "por", "pt", "português", "Português", "PORTUGUESE"],
    flag: "🇵🇹",
  },
  {
    names: ["japanese", "jpn", "ja", "日本語", "Japanese", "JAPANESE"],
    flag: "🇯🇵",
  },
  {
    names: [
      "chinese",
      "chi",
      "zh",
      "中文",
      "Chinese",
      "CHINESE",
      "mandarin",
      "Mandarin",
    ],
    flag: "🇨🇳",
  },
  {
    names: ["russian", "rus", "ru", "русский", "Russian", "RUSSIAN"],
    flag: "🇷🇺",
  },
  { names: ["korean", "kor", "ko", "한국어", "Korean", "KOREAN"], flag: "🇰🇷" },
  { names: ["arabic", "ara", "ar", "العربية", "Arabic", "ARABIC"], flag: "🇸🇦" },
  { names: ["hindi", "hin", "hi", "हिन्दी", "Hindi", "HINDI"], flag: "🇮🇳" },
  {
    names: ["bengali", "ben", "bn", "বাংলা", "Bengali", "BENGALI"],
    flag: "🇧🇩",
  },
  { names: ["urdu", "urd", "ur", "اردو", "Urdu", "URDU"], flag: "🇵🇰" },
  {
    names: ["turkish", "tur", "tr", "Türkçe", "Turkish", "TURKISH"],
    flag: "🇹🇷",
  },
  {
    names: ["persian", "fas", "fa", "فارسی", "Persian", "Farsi", "PERSIAN"],
    flag: "🇮🇷",
  },
  {
    names: ["swahili", "swa", "sw", "Kiswahili", "Swahili", "SWAHILI"],
    flag: "🇰🇪",
  },
  { names: ["dutch", "nld", "nl", "Nederlands", "Dutch", "DUTCH"], flag: "🇳🇱" },
  { names: ["polish", "pol", "pl", "Polski", "Polish", "POLISH"], flag: "🇵🇱" },
  {
    names: ["ukrainian", "ukr", "uk", "українська", "Ukrainian", "UKRAINIAN"],
    flag: "🇺🇦",
  },
  { names: ["greek", "ell", "el", "Ελληνικά", "Greek", "GREEK"], flag: "🇬🇷" },
  { names: ["hebrew", "heb", "he", "עברית", "Hebrew", "HEBREW"], flag: "🇮🇱" },
  { names: ["thai", "tha", "th", "ไทย", "Thai", "THAI"], flag: "🇹🇭" },
  {
    names: [
      "vietnamese",
      "vie",
      "vi",
      "Tiếng Việt",
      "Vietnamese",
      "VIETNAMESE",
    ],
    flag: "🇻🇳",
  },
  {
    names: ["romanian", "ron", "ro", "Română", "Romanian", "ROMANIAN"],
    flag: "🇷🇴",
  },
  {
    names: ["hungarian", "hun", "hu", "Magyar", "Hungarian", "HUNGARIAN"],
    flag: "🇭🇺",
  },
  { names: ["czech", "ces", "cs", "Čeština", "Czech", "CZECH"], flag: "🇨🇿" },
  {
    names: ["slovak", "slk", "sk", "Slovenčina", "Slovak", "SLOVAK"],
    flag: "🇸🇰",
  },
  {
    names: ["croatian", "hrv", "hr", "Hrvatski", "Croatian", "CROATIAN"],
    flag: "🇭🇷",
  },
  {
    names: ["serbian", "srp", "sr", "Српски", "Serbian", "SERBIAN"],
    flag: "🇷🇸",
  },
  {
    names: ["bulgarian", "bul", "bg", "Български", "Bulgarian", "BULGARIAN"],
    flag: "🇧🇬",
  },
  {
    names: ["finnish", "fin", "fi", "Suomi", "Finnish", "FINNISH"],
    flag: "🇫🇮",
  },
  {
    names: ["swedish", "swe", "sv", "Svenska", "Swedish", "SWEDISH"],
    flag: "🇸🇪",
  },
  {
    names: ["norwegian", "nor", "no", "Norsk", "Norwegian", "NORWEGIAN"],
    flag: "🇳🇴",
  },
  { names: ["danish", "dan", "da", "Dansk", "Danish", "DANISH"], flag: "🇩🇰" },
  {
    names: ["estonian", "est", "et", "Eesti", "Estonian", "ESTONIAN"],
    flag: "🇪🇪",
  },
  {
    names: ["latvian", "lav", "lv", "Latviešu", "Latvian", "LATVIAN"],
    flag: "🇱🇻",
  },
  {
    names: ["lithuanian", "lit", "lt", "Lietuvių", "Lithuanian", "LITHUANIAN"],
    flag: "🇱🇹",
  },
  {
    names: ["slovenian", "slv", "sl", "Slovenščina", "Slovenian", "SLOVENIAN"],
    flag: "🇸🇮",
  },
  {
    names: ["filipino", "fil", "tl", "Tagalog", "Filipino", "FILIPINO"],
    flag: "🇵🇭",
  },
  {
    names: ["malay", "msa", "ms", "Bahasa Melayu", "Malay", "MALAY"],
    flag: "🇲🇾",
  },
  {
    names: [
      "indonesian",
      "ind",
      "id",
      "Bahasa Indonesia",
      "Indonesian",
      "INDONESIAN",
    ],
    flag: "🇮🇩",
  },
  {
    names: ["burmese", "mya", "my", "မြန်မာ", "Burmese", "BURMESE"],
    flag: "🇲🇲",
  },
  { names: ["lao", "lao", "lo", "ລາວ", "Lao", "LAO"], flag: "🇱🇦" },
  { names: ["khmer", "khm", "km", "ភាសាខ្មែរ", "Khmer", "KHMER"], flag: "🇰🇭" },
  {
    names: ["mongolian", "mon", "mn", "Монгол", "Mongolian", "MONGOLIAN"],
    flag: "🇲🇳",
  },
  { names: ["nepali", "nep", "ne", "नेपाली", "Nepali", "NEPALI"], flag: "🇳🇵" },
  {
    names: ["sinhala", "sin", "si", "සිංහල", "Sinhala", "SINHALA"],
    flag: "🇱🇰",
  },
  { names: ["tamil", "tam", "ta", "தமிழ்", "Tamil", "TAMIL"], flag: "🇮🇳" },
  { names: ["telugu", "tel", "te", "తెలుగు", "Telugu", "TELUGU"], flag: "🇮🇳" },
  {
    names: ["marathi", "mar", "mr", "मराठी", "Marathi", "MARATHI"],
    flag: "🇮🇳",
  },
  {
    names: ["gujarati", "guj", "gu", "ગુજરાતી", "Gujarati", "GUJARATI"],
    flag: "🇮🇳",
  },
  {
    names: ["punjabi", "pan", "pa", "ਪੰਜਾਬੀ", "Punjabi", "PUNJABI"],
    flag: "🇮🇳",
  },
  { names: ["yoruba", "yor", "yo", "Yorùbá", "Yoruba", "YORUBA"], flag: "🇳🇬" },
  { names: ["igbo", "ibo", "ig", "Igbo", "IGBO"], flag: "🇳🇬" },
  { names: ["hausa", "hau", "ha", "Hausa", "HAUSA"], flag: "🇳🇬" },
  { names: ["zulu", "zul", "zu", "isiZulu", "Zulu", "ZULU"], flag: "🇿🇦" },
  { names: ["afrikaans", "afr", "af", "Afrikaans", "AFRIKAANS"], flag: "🇿🇦" },
  { names: ["amharic", "amh", "am", "አማርኛ", "Amharic", "AMHARIC"], flag: "🇪🇹" },
  {
    names: ["somali", "som", "so", "Soomaaliga", "Somali", "SOMALI"],
    flag: "🇸🇴",
  },
  { names: ["pashto", "pus", "ps", "پښتو", "Pashto", "PASHTO"], flag: "🇦🇫" },
  { names: ["uzbek", "uzb", "uz", "Oʻzbek", "Uzbek", "UZBEK"], flag: "🇺🇿" },
  { names: ["kazakh", "kaz", "kk", "Қазақ", "Kazakh", "KAZAKH"], flag: "🇰🇿" },
  { names: ["tajik", "tgk", "tg", "Тоҷикӣ", "Tajik", "TAJIK"], flag: "🇹🇯" },
  {
    names: ["georgian", "kat", "ka", "ქართული", "Georgian", "GEORGIAN"],
    flag: "🇬🇪",
  },
  {
    names: ["armenian", "hye", "hy", "Հայերեն", "Armenian", "ARMENIAN"],
    flag: "🇦🇲",
  },
  {
    names: [
      "azerbaijani",
      "aze",
      "az",
      "Azərbaycan",
      "Azerbaijani",
      "AZERBAIJANI",
    ],
    flag: "🇦🇿",
  },
  {
    names: ["albanian", "sqi", "sq", "Shqip", "Albanian", "ALBANIAN"],
    flag: "🇦🇱",
  },
  {
    names: ["bosnian", "bos", "bs", "Bosanski", "Bosnian", "BOSNIAN"],
    flag: "🇧🇦",
  },
  {
    names: [
      "macedonian",
      "mkd",
      "mk",
      "Македонски",
      "Macedonian",
      "MACEDONIAN",
    ],
    flag: "🇲🇰",
  },
  {
    names: [
      "belarusian",
      "bel",
      "be",
      "Беларуская",
      "Belarusian",
      "BELARUSIAN",
    ],
    flag: "🇧🇾",
  },
  {
    names: ["icelandic", "isl", "is", "Íslenska", "Icelandic", "ICELANDIC"],
    flag: "🇮🇸",
  },
  { names: ["irish", "gle", "ga", "Gaeilge", "Irish", "IRISH"], flag: "🇮🇪" },
  {
    names: [
      "scottish gaelic",
      "gla",
      "gd",
      "Gàidhlig",
      "Scottish Gaelic",
      "SCOTTISH GAELIC",
    ],
    flag: "🏴",
  },
  { names: ["welsh", "cym", "cy", "Cymraeg", "Welsh", "WELSH"], flag: "🏴" },
  { names: ["maori", "mri", "mi", "Māori", "Maori", "MAORI"], flag: "🇳🇿" },
  {
    names: ["samoan", "smo", "sm", "Gagana Samoa", "Samoan", "SAMOAN"],
    flag: "🇼🇸",
  },
  {
    names: ["tongan", "ton", "to", "Faka Tonga", "Tongan", "TONGAN"],
    flag: "🇹🇴",
  },
  { names: ["hawaiian", "haw", "hawai'i", "Hawaiian", "HAWAIIAN"], flag: "🇺🇸" },
  // Add more as needed
];

export function getLanguageFlag(language) {
  const normalized = String(language).trim().toLowerCase();
  for (const entry of languageFlagMap) {
    if (entry.names.some((name) => name.toLowerCase() === normalized)) {
      return entry.flag;
    }
  }
  return "🏳️"; // Default flag
}
