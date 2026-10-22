'use client';

import React from 'react';
import { useError } from './ErrorContext';
import ErrorToast from './ErrorToast';

export default function ToastContainer() {
  const { errors, removeError } = useError();
  console.log('[ToastContainer] Rendering with errors array:', errors);

  return (
    <div
      style={{
        position: 'fixed',
        top: '20px',
        right: '20px',
        zIndex: 9999,
        display: 'flex',
        flexDirection: 'column',
        gap: '10px',
        maxWidth: '440px',
        width: 'calc(100% - 40px)',
      }}
    >
      {errors.map((error: { id: string; message: string; data?: any }) => (
        <ErrorToast
          key={error.id}
          message={error.message}
          data={error.data}
          onDismiss={() => removeError(error.id)}
        />
      ))}
    </div>
  );
}
