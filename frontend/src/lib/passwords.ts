export const PASSWORD_ALGORITHMS = {
  argon2: {
    label: "Argon2id",
    description:
      "The current recommended algorithm, if unsure choose this algorithm.",
  },
  bcrypt: {
    label: "bcrypt",
    description:
      "Heavily supported and still secure, however new projects should migrate to Argon2id",
  },
} as const;
