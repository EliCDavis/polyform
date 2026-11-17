import { useEffect, useState } from "react";
import { useEditorOptional } from "../editor/EditorContext";

enum Support {
  Checking = "checking",
  Supported = "supported",
  Unsupported = "unsupported",
  Insecure = "insecure",
}

const SUPPORT_MESSAGES: Record<Exclude<Support, Support.Supported>, string> = {
  [Support.Checking]: "Checking for a headset...",
  [Support.Unsupported]: "No VR headset is available in this browser.",
  [Support.Insecure]: "VR needs the editor served over HTTPS or localhost.",
};

const SESSION_INIT: XRSessionInit = {
  optionalFeatures: ["local-floor", "bounded-floor", "layers"],
};

export function XRSection() {
  const renderer = useEditorOptional()?.threeApp.Renderer;
  const [support, setSupport] = useState(Support.Checking);
  const [presenting, setPresenting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!window.isSecureContext) {
      setSupport(Support.Insecure);
      return;
    }
    if (!navigator.xr) {
      setSupport(Support.Unsupported);
      return;
    }
    let cancelled = false;
    navigator.xr
      .isSessionSupported("immersive-vr")
      .then((ok) => !cancelled && setSupport(ok ? Support.Supported : Support.Unsupported))
      .catch(() => !cancelled && setSupport(Support.Unsupported));
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (!renderer) return;
    const started = () => setPresenting(true);
    const ended = () => setPresenting(false);
    setPresenting(renderer.xr.isPresenting);
    renderer.xr.addEventListener("sessionstart", started);
    renderer.xr.addEventListener("sessionend", ended);
    return () => {
      renderer.xr.removeEventListener("sessionstart", started);
      renderer.xr.removeEventListener("sessionend", ended);
    };
  }, [renderer]);

  const enter = async () => {
    if (!renderer || !navigator.xr) return;
    setError(null);
    try {
      const session = await navigator.xr.requestSession("immersive-vr", SESSION_INIT);
      await renderer.xr.setSession(session);
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e));
    }
  };

  const exit = () => {
    void renderer?.xr.getSession()?.end();
  };

  return (
    <>
      <div className="sidebar-header">XR</div>
      <div className="sidebar-section-content">
        {support !== Support.Supported && <div className="variable-description">{SUPPORT_MESSAGES[support]}</div>}
        {support === Support.Supported && !presenting && (
          <button type="button" onClick={enter} disabled={!renderer}>
            Enter VR
          </button>
        )}
        {presenting && (
          <button type="button" onClick={exit}>
            Exit VR
          </button>
        )}
        {error && <div className="variable-description">{error}</div>}
      </div>
    </>
  );
}
