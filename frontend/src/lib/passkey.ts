import * as star from "$lib/star";
import {
  sendSignal,
  startAuthentication,
  startRegistration,
} from "@simplewebauthn/browser";

const ceremony = async <T>(
  run: () => Promise<T>,
): Promise<T | star.StarError | null> => {
  try {
    return await run();
  } catch (e) {
    const name = e instanceof Error ? e.name : "";
    if (name === "NotAllowedError" || name === "AbortError") return null;

    return {
      error: true,
      status: 0,
      message: e instanceof Error ? e.message : "Your passkey request failed",
    };
  }
};

export const authenticateWithPasskey = async (
  flow?: string,
  useBrowserAutofill: boolean = false,
) => {
  const options = await star.getPasskeyLoginOptions(flow);
  if ("error" in options) return options;

  const credential = await ceremony(() =>
    startAuthentication({ optionsJSON: options.options, useBrowserAutofill }),
  );

  if (credential === null || "error" in credential) return credential;

  const response = await star.loginPasskey(options.challenge, credential, flow);

  if ("error" in response && response.status === 404 && options.options.rpId) {
    sendSignal({
      signalName: "unknownCredential",
      rpID: options.options.rpId,
      credentialID: credential.id,
    }).catch(() => {});
  }

  return response;
};

export const registerPasskey = async (name: string) => {
  const options = await star.enrollPasskey();
  if ("error" in options) return options;

  const credential = await ceremony(() =>
    startRegistration({ optionsJSON: options.options }),
  );

  if (credential === null || "error" in credential) return credential;
  return star.verifyPasskey(options.challenge, name, credential);
};

export const syncPasskeys = async () => {
  if (!("signalAllAcceptedCredentials" in PublicKeyCredential)) return;

  const signal = await star.getPasskeySignal();
  if (!signal) return;

  await Promise.allSettled([
    sendSignal({
      signalName: "allAcceptedCredentials",
      rpID: signal.rp_id,
      userID: signal.user_id,
      allAcceptedCredentialIDs: signal.credential_ids,
    }),
    sendSignal({
      signalName: "currentUserDetails",
      rpID: signal.rp_id,
      userID: signal.user_id,
      userName: signal.name,
      userDisplayName: signal.display_name,
    }),
  ]);
};
