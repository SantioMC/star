import type { SelectOption } from "@/components/select.svelte";
import type {
  AdminApplication,
  ApplicationConsent,
  SigningAlgorithm,
} from "@/star";

export type ClientKind = AdminApplication["kind"];

export const CLIENT_KIND_OPTIONS: SelectOption<ClientKind>[] = [
  { value: "public", label: "Public", hint: "PKCE" },
  { value: "confidential", label: "Confidential", hint: "client secret" },
];

export const CLIENT_KIND_DESCRIPTIONS: Record<ClientKind, string> = {
  public:
    "Browser or native apps that can not keep a secret. Authorization must use PKCE.",
  confidential:
    "Server-side apps that authenticate to the token endpoint with a client secret.",
};

export const CONSENT_OPTIONS: SelectOption<ApplicationConsent>[] = [
  { value: "default", label: "Default", hint: "once per scope set" },
  { value: "explicit", label: "Explicit", hint: "every sign in" },
  { value: "implicit", label: "Implicit", hint: "never" },
];

export const CONSENT_DESCRIPTIONS: Record<ApplicationConsent, string> = {
  default:
    "Users approve access the first time, and again only when the application requests new scopes.",
  explicit: "Users must approve access every time they sign in.",
  implicit:
    "Access is granted without asking. Only use this for trusted first-party applications.",
};

export const SIGNING_ALGORITHM_OPTIONS: SelectOption<SigningAlgorithm>[] = [
  { value: "ES512", label: "ES512", hint: "recommended" },
  { value: "RS256", label: "RS256", hint: "compatibility" },
];

export const SIGNING_ALGORITHM_DESCRIPTIONS: Record<SigningAlgorithm, string> =
  {
    ES512:
      "ECDSA with P-521. Compact and strong, supported by most modern OpenID Connect libraries.",
    RS256:
      "RSA with SHA-256. Use this for clients or resource servers that only support RS256.",
  };
