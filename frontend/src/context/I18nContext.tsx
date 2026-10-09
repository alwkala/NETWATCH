import React, { createContext, useContext, useEffect, useState, useMemo } from 'react';
import { Language } from '../types/settings';
import { en, Translations } from '../locales/en';
import { ar } from '../locales/ar';

interface I18nContextType {
  language: Language;
  setLanguage: (lang: Language) => void;
  dir: 'ltr' | 'rtl';
  isRTL: boolean;
  t: (path: string, params?: Record<string, string | number>) => string;
}

const dictionaries: Record<Language, Translations> = { en, ar };

const I18nContext = createContext<I18nContextType | undefined>(undefined);

export const I18nProvider: React.FC<{
  children: React.ReactNode;
  initialLanguage?: Language;
  onLanguageChange?: (lang: Language) => void;
}> = ({ children, initialLanguage, onLanguageChange }) => {
  const [language, setLanguageState] = useState<Language>(() => {
    if (initialLanguage === 'en' || initialLanguage === 'ar') {
      return initialLanguage;
    }
    const saved = localStorage.getItem('netwatch-lang');
    if (saved === 'en' || saved === 'ar') {
      return saved;
    }
    // Check browser navigator language
    if (typeof navigator !== 'undefined' && navigator.language?.startsWith('ar')) {
      return 'ar';
    }
    return 'en';
  });

  useEffect(() => {
    if (initialLanguage && (initialLanguage === 'en' || initialLanguage === 'ar') && initialLanguage !== language) {
      setLanguageState(initialLanguage);
    }
  }, [initialLanguage]);

  useEffect(() => {
    const root = document.documentElement;
    const isRtl = language === 'ar';
    root.dir = isRtl ? 'rtl' : 'ltr';
    root.lang = language;
    if (isRtl) {
      root.classList.add('rtl');
    } else {
      root.classList.remove('rtl');
    }
    localStorage.setItem('netwatch-lang', language);
  }, [language]);

  const setLanguage = (newLang: Language) => {
    setLanguageState(newLang);
    onLanguageChange?.(newLang);
  };

  const isRTL = language === 'ar';
  const dir: 'ltr' | 'rtl' = isRTL ? 'rtl' : 'ltr';

  const t = useMemo(() => {
    return (path: string, params?: Record<string, string | number>): string => {
      const parts = path.split('.');
      const currentDict = dictionaries[language] || en;
      
      let val: any = currentDict;
      for (const part of parts) {
        if (val && typeof val === 'object' && part in val) {
          val = val[part];
        } else {
          val = undefined;
          break;
        }
      }

      // Fallback to English if value is missing
      if (val === undefined || typeof val !== 'string') {
        let fallbackVal: any = en;
        for (const part of parts) {
          if (fallbackVal && typeof fallbackVal === 'object' && part in fallbackVal) {
            fallbackVal = fallbackVal[part];
          } else {
            fallbackVal = undefined;
            break;
          }
        }
        val = typeof fallbackVal === 'string' ? fallbackVal : path;
      }

      let res = String(val);
      if (params) {
        for (const [k, v] of Object.entries(params)) {
          res = res.replaceAll(`{${k}}`, String(v));
        }
      }
      return res;
    };
  }, [language]);

  return (
    <I18nContext.Provider value={{ language, setLanguage, dir, isRTL, t }}>
      {children}
    </I18nContext.Provider>
  );
};

export const useI18n = (): I18nContextType => {
  const context = useContext(I18nContext);
  if (!context) {
    throw new Error('useI18n must be used within an I18nProvider');
  }
  return context;
};
