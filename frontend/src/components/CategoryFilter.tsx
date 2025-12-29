'use client'

interface CategoryFilterProps {
  selected: string
  onChange: (category: string) => void
}

const categories = [
  { value: 'all', label: 'All Events', icon: '🎉' },
  { value: 'music', label: 'Music', icon: '🎵' },
  { value: 'sports', label: 'Sports', icon: '⚽' },
  { value: 'conference', label: 'Conference', icon: '🎤' },
  { value: 'arts', label: 'Arts', icon: '🎨' },
  { value: 'food', label: 'Food & Drink', icon: '🍽️' },
  { value: 'festival', label: 'Festival', icon: '🎪' },
]

export default function CategoryFilter({ selected, onChange }: CategoryFilterProps) {
  return (
    <div className="flex gap-2 overflow-x-auto pb-2 scrollbar-hide">
      {categories.map((category) => (
        <button
          key={category.value}
          onClick={() => onChange(category.value)}
          className={`
            flex items-center gap-2 px-4 py-2 rounded-lg whitespace-nowrap transition-all
            ${selected === category.value
              ? 'bg-brand-600 text-white shadow-lg shadow-brand-600/30'
              : 'bg-white text-gray-700 hover:bg-gray-50 border border-gray-200'
            }
          `}
        >
          <span>{category.icon}</span>
          <span className="font-medium">{category.label}</span>
        </button>
      ))}
    </div>
  )
}
