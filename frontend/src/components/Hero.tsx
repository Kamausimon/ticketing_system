import Link from 'next/link'

export default function Hero() {
  return (
    <div className="relative bg-gradient-to-br from-brand-50 via-white to-accent-50 overflow-hidden">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24 lg:py-32">
        <div className="text-center">
          <h1 className="text-4xl sm:text-5xl lg:text-6xl font-bold text-gray-900 mb-6 text-balance">
            Discover Events That
            <span className="text-brand-600"> Move You</span>
          </h1>
          <p className="text-xl text-gray-600 mb-8 max-w-2xl mx-auto text-balance">
            From concerts to conferences, find and book tickets to the best events happening around you. Secure, instant, and hassle-free.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link
              href="/events"
              className="bg-brand-600 text-white px-8 py-3 rounded-lg text-lg font-semibold hover:bg-brand-700 transition-colors shadow-lg shadow-brand-600/30"
            >
              Explore Events
            </Link>
            <Link
              href="/create-event"
              className="bg-white text-brand-600 px-8 py-3 rounded-lg text-lg font-semibold hover:bg-gray-50 transition-colors border-2 border-brand-600"
            >
              Host an Event
            </Link>
          </div>

          {/* Stats */}
          <div className="mt-16 grid grid-cols-3 gap-8 max-w-3xl mx-auto">
            <div>
              <div className="text-3xl font-bold text-brand-600">50K+</div>
              <div className="text-sm text-gray-600 mt-1">Active Events</div>
            </div>
            <div>
              <div className="text-3xl font-bold text-brand-600">1M+</div>
              <div className="text-sm text-gray-600 mt-1">Tickets Sold</div>
            </div>
            <div>
              <div className="text-3xl font-bold text-brand-600">10K+</div>
              <div className="text-sm text-gray-600 mt-1">Organizers</div>
            </div>
          </div>
        </div>
      </div>

      {/* Decorative elements */}
      <div className="absolute top-0 right-0 -translate-y-12 translate-x-12 w-96 h-96 bg-accent-200 rounded-full opacity-20 blur-3xl" />
      <div className="absolute bottom-0 left-0 translate-y-12 -translate-x-12 w-96 h-96 bg-brand-200 rounded-full opacity-20 blur-3xl" />
    </div>
  )
}
