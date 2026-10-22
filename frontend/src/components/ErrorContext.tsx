'use client';

import React, { useEffect, useState, createContext, useContext } from 'react';
import ErrorAlert from '@/components/ErrorAlert';

interface ErrorItem {
  id: string;
  message: string;
  data?: any;
}

interface ErrorContextType {
  errors: ErrorItem[];
  addError: (message: string, data?: any) => void;
  removeError: (id: string) => void;
  clearErrors: () => void;
}

const ErrorContext = createContext<ErrorContextType | undefined>(undefined);

export function ErrorProvider({ children }: { children: React.ReactNode }) {
  const [errors, setErrors] = useState<ErrorItem[]>([]);

  const addError = (message: string, data?: any) => {
    console.log('[ErrorContext] addError invoked', { message, data });
    const id = Math.random().toString(36).substring(7);
    setErrors((prev) => {
      const newErrors = [...prev, { id, message, data }];
      console.log('[ErrorContext] New error queue:', newErrors);
      return newErrors;
    });
  };

  const removeError = (id: string) => {
    setErrors((prev) => prev.filter((err) => err.id !== id));
  };

  const clearErrors = () => {
    setErrors([]);
  };

  return (
    <ErrorContext.Provider value={{ errors, addError, removeError, clearErrors }}>
      {children}
    </ErrorContext.Provider>
  );
}

export function useError() {
  const context = useContext(ErrorContext);
  if (!context) {
    throw new Error('useError must be used within an ErrorProvider');
  }
  return context;
}
