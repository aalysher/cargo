/** Профиль: данные, телефон, адрес склада, правила, поддержка. */
import { useState } from 'react';
import { ChevronRight, LifeBuoy, MapPin, Phone, ShieldCheck } from 'lucide-react';
import { api, type ApiError } from '../lib/api';
import type { Profile } from '../lib/types';
import { BOT_USERNAME } from '../lib/config';
import { openBotChat } from '../lib/telegram';
import { toast } from '../components/Toast';
import { Button, Card } from '../components/primitives';

export function ProfilePage({
  profile,
  onProfileChange,
  onOpenAddress,
}: {
  profile: Profile;
  onProfileChange: (p: Profile) => void;
  onOpenAddress: () => void;
}) {
  const [phone, setPhone] = useState(profile.phone);
  const [saving, setSaving] = useState(false);

  const savePhone = async () => {
    if (saving) return;
    setSaving(true);
    try {
      await api('/me', { method: 'PATCH', body: { phone } });
      toast('Телефон сохранён');
      onProfileChange({ ...profile, phone });
    } catch (e) {
      toast((e as ApiError).detail, 'err');
    } finally {
      setSaving(false);
    }
  };

  const name = [profile.first_name, profile.last_name].filter(Boolean).join(' ') || 'Клиент';

  return (
    <div className="space-y-3 px-4 pb-28 pt-4">
      {/* Карточка профиля */}
      <Card className="p-4">
        <div className="flex items-center gap-3">
          {profile.photo_url ? (
            <img src={profile.photo_url} alt="" className="h-14 w-14 rounded-full object-cover" />
          ) : (
            <div className="flex h-14 w-14 items-center justify-center rounded-full bg-brand/10 text-h1 text-brand">
              {name[0]}
            </div>
          )}
          <div>
            <div className="text-h2">{name}</div>
            <div className="text-caption text-ink-muted">
              {profile.username ? `@${profile.username} · ` : ''}код {profile.client_code}
            </div>
          </div>
        </div>
      </Card>

      {/* Телефон */}
      <Card className="p-4">
        <div className="flex items-center gap-2 text-h2">
          <Phone size={18} className="text-brand" /> Телефон
        </div>
        <p className="mt-1 text-caption text-ink-muted">Нужен для связи при выдаче посылки</p>
        <input
          className="field tnum mt-2"
          inputMode="tel"
          placeholder="+996555123456"
          value={phone}
          onChange={(e) => setPhone(e.target.value)}
        />
        <Button variant="secondary" className="mt-2 h-11" onClick={savePhone} disabled={saving}>
          Сохранить
        </Button>
      </Card>

      {/* Ссылки */}
      <Card>
        <MenuRow icon={<MapPin size={18} className="text-brand" />} title="Мой адрес склада" onClick={onOpenAddress} />
        <MenuRow
          icon={<LifeBuoy size={18} className="text-brand" />}
          title="Поддержка (чат с ботом)"
          onClick={() => openBotChat(BOT_USERNAME)}
          last
        />
      </Card>

      {/* Правила */}
      <Card className="p-4">
        <div className="flex items-center gap-2 text-h2">
          <ShieldCheck size={18} className="text-brand" /> Правила перевозки
        </div>
        <div className="mt-2 space-y-1.5 text-caption text-ink-muted">
          <p>✅ Можно: одежда, обувь, электроника, товары для дома, аксессуары.</p>
          <p>❌ Нельзя: запрещённые законом грузы, лекарства без документов, жидкости свыше лимитов перевозчика.</p>
          <p>📦 Хрупкое и дорогое — отметьте в комментарии к заявке, оператор уточнит упаковку.</p>
          <p>Сомневаетесь, можно ли везти товар? Напишите в поддержку — ответим по конкретной позиции.</p>
        </div>
      </Card>
    </div>
  );
}

function MenuRow({ icon, title, onClick, last = false }: { icon: React.ReactNode; title: string; onClick: () => void; last?: boolean }) {
  return (
    <button
      onClick={onClick}
      className={`flex w-full items-center justify-between px-4 py-3.5 text-left transition-colors active:bg-surface-2 ${
        last ? '' : 'border-b border-line'
      }`}
    >
      <span className="flex items-center gap-3 text-body">
        {icon} {title}
      </span>
      <ChevronRight size={18} className="text-ink-muted" />
    </button>
  );
}
