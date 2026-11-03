export interface UsernameRules {
  min_length: number;
  max_length: number;
  allow_dash: boolean;
  allow_dots: boolean;
  allow_underscore: boolean;
  allow_spaces: boolean;
}

export const validateUsername = (
  input: string,
  rules: UsernameRules,
): string | null => {
  const username = input.trim();

  if (username.length < rules.min_length) {
    return `Username must be at least ${rules.min_length} characters`;
  }

  if (username.length > rules.max_length) {
    return `Username must be at most ${rules.max_length} characters`;
  }

  if (!/^[A-Za-z0-9._\- ]+$/.test(username)) {
    return "Username may only contain English letters, numbers, dots, dashes, underscores and spaces";
  }

  if (!rules.allow_spaces && username.includes(" ")) {
    return "Username must not contain a space";
  }

  if (!rules.allow_dash && username.includes("-")) {
    return "Username must not contain a dash";
  } else if (username.startsWith("-") || username.endsWith("-")) {
    return "Username must not start or end with a dash";
  } else if (username.includes("--")) {
    return "Username must not contain consecutive dashes";
  }

  if (!rules.allow_dots && username.includes(".")) {
    return "Username must not contain a dot";
  } else if (username.includes("..")) {
    return "Username must not contain consecutive dots";
  }

  if (!rules.allow_underscore && username.includes("_")) {
    return "Username must not contain an underscore";
  }

  if (!/[a-zA-Z]/.test(username)) {
    return "Username must contain at least one letter";
  }

  return null;
};
