// Ports of quaternion.FromEulerAngle and ToEulerAngles so the editor agrees
// with the backend on which way a rotation turns.

export interface Quat {
  x: number;
  y: number;
  z: number;
  w: number;
}

export interface EulerDegrees {
  x: number;
  y: number;
  z: number;
}

export const identityQuat: Quat = { x: 0, y: 0, z: 0, w: 1 };

export function eulerToQuat(e: EulerDegrees): Quat {
  const rad = Math.PI / 180;
  const cr = Math.cos(e.x * rad * 0.5);
  const sr = Math.sin(e.x * rad * 0.5);
  const cp = Math.cos(e.y * rad * 0.5);
  const sp = Math.sin(e.y * rad * 0.5);
  const cy = Math.cos(e.z * rad * 0.5);
  const sy = Math.sin(e.z * rad * 0.5);
  return {
    x: sr * cp * cy - cr * sp * sy,
    y: cr * sp * cy + sr * cp * sy,
    z: cr * cp * sy - sr * sp * cy,
    w: cr * cp * cy + sr * sp * sy,
  };
}

export function quatToEuler(q: Quat): EulerDegrees {
  const deg = 180 / Math.PI;
  const x = Math.atan2(2 * (q.w * q.x + q.y * q.z), 1 - 2 * (q.x * q.x + q.y * q.y));
  const sinp = Math.sqrt(1 + 2 * (q.w * q.y - q.x * q.z));
  const cosp = Math.sqrt(1 - 2 * (q.w * q.y - q.x * q.z));
  const y = 2 * Math.atan2(sinp, cosp) - Math.PI / 2;
  const z = Math.atan2(2 * (q.w * q.z + q.x * q.y), 1 - 2 * (q.y * q.y + q.z * q.z));
  return { x: x * deg, y: y * deg, z: z * deg };
}

export function quatEquals(a: Quat, b: Quat, epsilon = 1e-9): boolean {
  return (
    Math.abs(a.x - b.x) < epsilon &&
    Math.abs(a.y - b.y) < epsilon &&
    Math.abs(a.z - b.z) < epsilon &&
    Math.abs(a.w - b.w) < epsilon
  );
}
