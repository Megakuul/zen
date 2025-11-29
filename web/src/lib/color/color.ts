/**
 * GetScoreDecorator returns a text decorator for the provided score
 */
export function GetScoreDecorator(score: number): string {
  switch (true) {
    case score <= -1000:
      return 'text-red-950/90';
    case score <= -500:
      return 'text-red-800/90';
    case score <= -100:
      return 'text-red-200/80';
    case score < 0:
      return 'text-orange-700/80';
    case score == 0:
      return 'text-slate-50/90';
    case score >= 1000:
      return 'text-black [-webkit-text-stroke:1px_white] [text-shadow:0_0_4px_white,0_0_6px_white,0_0_12px_white]';
    case score >= 500:
      return 'text-emerald-700/90';
    case score >= 100:
      return 'text-emerald-400/80';
    case score > 0:
      return 'text-emerald-200/80';
    default:
      return '';
  }
}

/**
 * GetChangeDecorator returns a text decorator for the provided rating change
 */
export function GetChangeDecorator(change: number): string {
  switch (true) {
    case change <= -100:
      return 'text-red-950/90';
    case change <= -50:
      return 'text-red-800/90';
    case change <= -10:
      return 'text-red-200/80';
    case change < 0:
      return 'text-orange-700/80';
    case change == 0:
      return 'text-slate-50/90';
    case change >= 100:
      return 'text-black [-webkit-text-stroke:1px_white] [text-shadow:0_0_4px_white,0_0_6px_white,0_0_12px_white]';
    case change >= 50:
      return 'text-emerald-700/90';
    case change >= 10:
      return 'text-emerald-400/80';
    case change > 0:
      return 'text-emerald-200/80';
    default:
      return '';
  }
}
