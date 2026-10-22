export interface ParsedError {
  message: string;
  code?: string;
  fields?: Array<{ field: string; message: string }>;
  data?: any;
}

/**
   * Parses an appError or any nested stringified JSON error structure recursively.
   * Handles cases where message is a JSON string, contains nested JSON, or follows the backend Response format.
   */
export function parseAppError(input: any): ParsedError {
  let message = 'An unexpected error occurred';
  let code: string | undefined = undefined;
  let fields: Array<{ field: string; message: string }> | undefined = undefined;
  let data: any = undefined;

  if (!input) {
    return { message };
  }

  // If it's an Error instance or object with message/data/details
  let current: any = input;
  if (input instanceof Error) {
    message = input.message;
    if ((input as any).details) {
      data = (input as any).details;
    }
  } else if (typeof input === 'string') {
    message = input;
  } else if (typeof input === 'object') {
    if (input.message) message = String(input.message);
    if (input.code) code = String(input.code);
    if (input.data) data = input.data;
    if (input.fields) fields = input.fields;
  }

  // Check if message is a JSON string (possibly multiple times or nested)
  let maxDepth = 3;
  while (maxDepth > 0 && typeof message === 'string') {
    const trimmed = message.trim();
    if ((trimmed.startsWith('{') && trimmed.endsWith('}')) || (trimmed.startsWith('[') && trimmed.endsWith(']'))) {
      try {
        const parsed = JSON.parse(trimmed);
        if (parsed) {
          if (parsed.message) {
            message = String(parsed.message);
          }
          if (parsed.code) {
            code = String(parsed.code);
          }
          if (parsed.meta && parsed.meta.code) {
            code = String(parsed.meta.code);
          }
          if (parsed.data) {
            data = parsed.data;
            if (Array.isArray(parsed.data)) {
              fields = parsed.data;
            }
          }
          if (parsed.fields) {
            fields = parsed.fields;
          }
          maxDepth--;
          continue;
        }
      } catch (e) {
        // Not valid JSON, break loop
      }
    }
    break;
  }

  // Also if data contains fields or is an array
  if (data && !fields && Array.isArray(data)) {
    fields = data;
  }

  // Clean up any remaining prefix like "EXPORT_LIMIT_EXCEEDED: " if code wasn't set
  if (!code && typeof message === 'string' && message.includes(': ')) {
    const parts = message.split(': ');
    if (parts.length >= 2 && parts[0].toUpperCase() === parts[0] && parts[0].length < 30) {
      code = parts[0];
      message = parts.slice(1).join(': ');
    }
  }

  return {
    message,
    code,
    fields,
    data,
  };
}
