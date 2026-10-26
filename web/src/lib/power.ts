export const parseQuantity = (text: string) => {
  const trimmed = text.trim();
  const value = trimmed === "" ? Number.NaN : Number(trimmed);
  return Number.isFinite(value) ? value : Number.NaN;
};

export const parsePowerFactor = (text: string) => {
  const value = parseQuantity(text);
  return value > 0 && value <= 1 ? value : 1;
};

export const solvePower = (
  powerText: string,
  voltageText: string,
  currentText: string,
  powerFactorText = "",
) => {
  const power = parseQuantity(powerText);
  const voltage = parseQuantity(voltageText);
  const current = parseQuantity(currentText);
  const powerFactor = parsePowerFactor(powerFactorText);

  return {
    power: Number.isNaN(power) ? voltage * current * powerFactor : power,
    voltage: Number.isNaN(voltage)
      ? current > 0
        ? power / (current * powerFactor)
        : Number.NaN
      : voltage,
    current: Number.isNaN(current)
      ? voltage > 0
        ? power / (voltage * powerFactor)
        : Number.NaN
      : current,
  };
};
