/** Онбординг (1-й запуск): 3 шага + MainButton (вне Telegram — инлайн-кнопка). */
import { useState } from 'react';
import { CheckCircle2, Package, Truck } from 'lucide-react';
import { useMainButton } from '../hooks/useTelegramUI';
import { hapticNotify, isTMA } from '../lib/telegram';
import { Button } from '../components/primitives';

const STEPS = [
  {
    icon: Package,
    title: '1. Покупаете',
    text: 'Заказываете на Taobao, Poizon, 1688 — что угодно из Китая.',
  },
  {
    icon: Truck,
    title: '2. Отправляете на наш склад',
    text: 'Адрес склада и ваш личный код — всегда под рукой в приложении, копируются в один тап.',
  },
  {
    icon: CheckCircle2,
    title: '3. Получаете в Бишкеке',
    text: 'Следите за статусами, мы сообщим, когда посылка будет готова к выдаче.',
  },
];

export function Onboarding({ onFinish }: { onFinish: () => void }) {
  const [step, setStep] = useState(0);
  const isLast = step === STEPS.length - 1;
  const next = () => {
    hapticNotify('success');
    if (isLast) onFinish();
    else setStep((s) => s + 1);
  };

  useMainButton(isLast ? 'Начать' : 'Далее', next);

  const { icon: Icon, title, text } = STEPS[step];

  return (
    <div className="flex min-h-dvh flex-col items-center justify-center px-6 pb-16">
      <div className="mb-10 flex h-24 w-24 items-center justify-center rounded-full bg-brand/10">
        <Icon size={44} strokeWidth={1.75} className="text-brand" />
      </div>
      <div key={step} className="animate-fade-up text-center">
        <div className="text-h1">{title}</div>
        <p className="mx-auto mt-3 max-w-[300px] text-body text-ink-muted">{text}</p>
      </div>
      {!isTMA() && (
        <div className="absolute bottom-12 w-full max-w-[280px] px-6">
          <Button onClick={next}>{isLast ? 'Начать' : 'Далее'}</Button>
        </div>
      )}
      <div className="absolute bottom-6 flex gap-1.5">
        {STEPS.map((_, i) => (
          <span key={i} className={`h-1.5 w-1.5 rounded-full ${i === step ? 'bg-brand' : 'bg-line'}`} />
        ))}
      </div>
    </div>
  );
}
