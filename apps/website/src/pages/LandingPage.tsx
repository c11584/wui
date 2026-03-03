import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { 
  Rocket, 
  Shield, 
  Zap, 
  Globe, 
  Users, 
  BarChart3,
  ChevronDown,
  Menu,
  X,
  Check,
  ArrowRight,
  Github,
  Twitter,
  Mail
} from 'lucide-react'

const fadeInUp = {
  initial: { opacity: 0, y: 20 },
  animate: { opacity: 1, y: 0 },
  transition: { duration: 0.6 }
}

export default function LandingPage() {
  const [scrollY, setScrollY] = useState(0)
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const [activeFaq, setActiveFaq] = useState<number | null>(null)

  useEffect(() => {
    const handleScroll = () => setScrollY(window.scrollY)
    window.addEventListener('scroll', handleScroll)
    return () => window.removeEventListener('scroll', handleScroll)
  }, [])

  const features = [
    {
      icon: Rocket,
      title: 'Tunnel Orchestration',
      description: 'Inbound protocol + port → Outbound protocol + port with UDP support',
      gradient: 'from-blue-500 to-cyan-500'
    },
    {
      icon: Shield,
      title: 'Enterprise Security',
      description: 'JWT authentication + 2FA support + encrypted data storage',
      gradient: 'from-purple-500 to-pink-500'
    },
    {
      icon: Zap,
      title: 'One-Click Install',
      description: 'Automated installation with configurable parameters',
      gradient: 'from-yellow-500 to-orange-500'
    },
    {
      icon: Globe,
      title: 'Multi-Protocol',
      description: 'VMess, VLESS, Trojan, Shadowsocks, Hysteria, SOCKS5, HTTP',
      gradient: 'from-green-500 to-teal-500'
    },
    {
      icon: Users,
      title: 'Multi-User Support',
      description: 'Complete user management with RBAC permissions',
      gradient: 'from-indigo-500 to-purple-500'
    },
    {
      icon: BarChart3,
      title: 'Real-time Monitoring',
      description: 'Traffic statistics + connection tracking + live dashboards',
      gradient: 'from-red-500 to-pink-500'
    }
  ]

  const pricingPlans = [
    {
      name: 'Personal',
      price: '$9',
      period: '/month',
      features: [
        'Up to 10 tunnels',
        '100GB traffic/month',
        'All protocols supported',
        'Email support',
        'Single user'
      ],
      cta: 'Get Started',
      popular: false
    },
    {
      name: 'Team',
      price: '$29',
      period: '/month',
      features: [
        'Up to 50 tunnels',
        '500GB traffic/month',
        'All protocols supported',
        'Priority support',
        'Up to 10 users',
        'API access'
      ],
      cta: 'Start Free Trial',
      popular: true
    },
    {
      name: 'Enterprise',
      price: 'Custom',
      period: '',
      features: [
        'Unlimited tunnels',
        'Unlimited traffic',
        'All protocols + custom',
        '24/7 dedicated support',
        'Unlimited users',
        'Cluster support',
        'SLA guarantee'
      ],
      cta: 'Contact Sales',
      popular: false
    }
  ]

  const faqs = [
    {
      q: 'What is WUI?',
      a: 'WUI is a next-generation proxy management panel that surpasses x-ui. It provides a modern web interface for managing Xray tunnels with multi-user support, real-time monitoring, and enterprise-grade security.'
    },
    {
      q: 'How does WUI compare to x-ui?',
      a: 'WUI offers a more modern UI with React 18, real-time WebSocket updates, multi-user support with RBAC, and commercial licensing options. It maintains compatibility with existing x-ui configurations while adding new features.'
    },
    {
      q: 'Can I try WUI before purchasing?',
      a: 'Yes! We offer a 14-day free trial for all plans. No credit card required. You can also explore our self-hosted version with basic features.'
    },
    {
      q: 'What protocols are supported?',
      a: 'WUI supports VMess, VLESS, Trojan, Shadowsocks, Hysteria, SOCKS5, and HTTP protocols for both inbound and outbound configurations.'
    },
    {
      q: 'Is WUI open source?',
      a: 'We offer both open-source and commercial versions. The core functionality is available under MIT license, while advanced features require a commercial license.'
    }
  ]

  return (
    <div className="min-h-screen text-white">
      {/* Navigation */}
      <nav className={`fixed top-0 left-0 right-0 z-50 transition-all duration-300 ${
        scrollY > 50 ? 'bg-gray-900/95 backdrop-blur-md shadow-lg' : ''
      }`}>
        <div className="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center font-bold text-lg">
              W
            </div>
            <span className="text-xl font-bold">WUI</span>
          </div>

          <div className="hidden md:flex items-center space-x-8">
            <a href="#features" className="text-gray-300 hover:text-white transition">Features</a>
            <a href="#pricing" className="text-gray-300 hover:text-white transition">Pricing</a>
            <a href="#faq" className="text-gray-300 hover:text-white transition">FAQ</a>
            <a href="https://github.com/your-org/wui" className="text-gray-300 hover:text-white transition flex items-center">
              <Github className="w-5 h-5 mr-1" />
              GitHub
            </a>
            <button className="px-6 py-2 bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg font-medium hover:opacity-90 transition">
              Get Started
            </button>
          </div>

          <button 
            className="md:hidden text-white"
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
          >
            {mobileMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
          </button>
        </div>

        {mobileMenuOpen && (
          <motion.div 
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            className="md:hidden bg-gray-900/95 backdrop-blur-md border-t border-gray-800"
          >
            <div className="px-6 py-4 space-y-4">
              <a href="#features" className="block text-gray-300 hover:text-white">Features</a>
              <a href="#pricing" className="block text-gray-300 hover:text-white">Pricing</a>
              <a href="#faq" className="block text-gray-300 hover:text-white">FAQ</a>
              <a href="https://github.com/your-org/wui" className="block text-gray-300 hover:text-white">GitHub</a>
              <button className="w-full px-6 py-2 bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg font-medium">
                Get Started
              </button>
            </div>
          </motion.div>
        )}
      </nav>

      {/* Hero Section */}
      <section className="pt-32 pb-20 px-6">
        <div className="max-w-7xl mx-auto">
          <motion.div 
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8 }}
            className="text-center"
          >
            <div className="inline-flex items-center px-4 py-2 rounded-full bg-blue-500/10 border border-blue-500/20 text-blue-400 text-sm mb-6">
              <Rocket className="w-4 h-4 mr-2" />
              Next-Generation Proxy Management
            </div>

            <h1 className="text-5xl md:text-7xl font-bold mb-6">
              <span className="gradient-text">Surpass x-ui</span>
              <br />
              <span className="text-white">with Modern Power</span>
            </h1>

            <p className="text-xl text-gray-400 max-w-3xl mx-auto mb-10">
              A high-performance proxy management panel with modern UI, multi-user support, 
              real-time monitoring, and enterprise-grade security.
            </p>

            <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
              <button className="px-8 py-4 bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg font-medium text-lg hover:opacity-90 transition flex items-center">
                Start Free Trial
                <ArrowRight className="w-5 h-5 ml-2" />
              </button>
              <button className="px-8 py-4 glass-card rounded-lg font-medium text-lg hover:bg-white/10 transition flex items-center">
                <Github className="w-5 h-5 mr-2" />
                View on GitHub
              </button>
            </div>

            <div className="mt-16 relative">
              <div className="absolute inset-0 bg-gradient-to-t from-gray-900 via-transparent to-transparent z-10" />
              <div className="glass-card rounded-2xl p-2 hover-lift">
                <img 
                  src="/dashboard-preview.png" 
                  alt="WUI Dashboard Preview" 
                  className="rounded-xl w-full shadow-2xl"
                />
              </div>
            </div>
          </motion.div>
        </div>
      </section>

      {/* Features Section */}
      <section id="features" className="py-20 px-6 bg-gray-900/50">
        <div className="max-w-7xl mx-auto">
          <motion.div 
            initial={{ opacity: 0 }}
            whileInView={{ opacity: 1 }}
            viewport={{ once: true }}
            className="text-center mb-16"
          >
            <h2 className="text-4xl font-bold mb-4">Powerful Features</h2>
            <p className="text-xl text-gray-400 max-w-2xl mx-auto">
              Everything you need to manage your proxy infrastructure
            </p>
          </motion.div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {features.map((feature, index) => (
              <motion.div
                key={feature.title}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ delay: index * 0.1 }}
                className="glass-card rounded-xl p-6 hover-lift"
              >
                <div className={`w-12 h-12 rounded-lg bg-gradient-to-br ${feature.gradient} flex items-center justify-center mb-4`}>
                  <feature.icon className="w-6 h-6 text-white" />
                </div>
                <h3 className="text-xl font-semibold mb-2">{feature.title}</h3>
                <p className="text-gray-400">{feature.description}</p>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* Pricing Section */}
      <section id="pricing" className="py-20 px-6">
        <div className="max-w-7xl mx-auto">
          <motion.div 
            initial={{ opacity: 0 }}
            whileInView={{ opacity: 1 }}
            viewport={{ once: true }}
            className="text-center mb-16"
          >
            <h2 className="text-4xl font-bold mb-4">Simple Pricing</h2>
            <p className="text-xl text-gray-400 max-w-2xl mx-auto">
              Choose the plan that fits your needs
            </p>
          </motion.div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {pricingPlans.map((plan, index) => (
              <motion.div
                key={plan.name}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ delay: index * 0.1 }}
                className={`relative glass-card rounded-xl p-8 hover-lift ${
                  plan.popular ? 'border-2 border-blue-500' : ''
                }`}
              >
                {plan.popular && (
                  <div className="absolute -top-4 left-1/2 transform -translate-x-1/2 px-4 py-1 bg-blue-500 rounded-full text-sm font-medium">
                    Most Popular
                  </div>
                )}

                <h3 className="text-2xl font-bold mb-4">{plan.name}</h3>
                <div className="mb-6">
                  <span className="text-4xl font-bold">{plan.price}</span>
                  <span className="text-gray-400">{plan.period}</span>
                </div>

                <ul className="space-y-3 mb-8">
                  {plan.features.map((feature) => (
                    <li key={feature} className="flex items-center">
                      <Check className="w-5 h-5 text-green-500 mr-3 flex-shrink-0" />
                      <span className="text-gray-300">{feature}</span>
                    </li>
                  ))}
                </ul>

                <button className={`w-full py-3 rounded-lg font-medium transition ${
                  plan.popular 
                    ? 'bg-gradient-to-r from-blue-500 to-purple-600 hover:opacity-90' 
                    : 'bg-gray-800 hover:bg-gray-700'
                }`}>
                  {plan.cta}
                </button>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* FAQ Section */}
      <section id="faq" className="py-20 px-6 bg-gray-900/50">
        <div className="max-w-3xl mx-auto">
          <motion.div 
            initial={{ opacity: 0 }}
            whileInView={{ opacity: 1 }}
            viewport={{ once: true }}
            className="text-center mb-16"
          >
            <h2 className="text-4xl font-bold mb-4">Frequently Asked Questions</h2>
            <p className="text-xl text-gray-400">
              Everything you need to know
            </p>
          </motion.div>

          <div className="space-y-4">
            {faqs.map((faq, index) => (
              <motion.div
                key={index}
                initial={{ opacity: 0 }}
                whileInView={{ opacity: 1 }}
                viewport={{ once: true }}
                transition={{ delay: index * 0.1 }}
                className="glass-card rounded-xl overflow-hidden"
              >
                <button
                  onClick={() => setActiveFaq(activeFaq === index ? null : index)}
                  className="w-full px-6 py-4 flex items-center justify-between text-left"
                >
                  <span className="font-medium">{faq.q}</span>
                  <ChevronDown className={`w-5 h-5 transition-transform ${
                    activeFaq === index ? 'rotate-180' : ''
                  }`} />
                </button>
                {activeFaq === index && (
                  <motion.div
                    initial={{ height: 0 }}
                    animate={{ height: 'auto' }}
                    className="overflow-hidden"
                  >
                    <p className="px-6 pb-4 text-gray-400">{faq.a}</p>
                  </motion.div>
                )}
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20 px-6">
        <div className="max-w-4xl mx-auto text-center">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
          >
            <h2 className="text-4xl font-bold mb-6">Ready to get started?</h2>
            <p className="text-xl text-gray-400 mb-10">
              Join thousands of users who trust WUI for their proxy management needs.
            </p>
            <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
              <button className="px-8 py-4 bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg font-medium text-lg hover:opacity-90 transition">
                Start Your Free Trial
              </button>
              <button className="px-8 py-4 glass-card rounded-lg font-medium text-lg hover:bg-white/10 transition flex items-center">
                <Mail className="w-5 h-5 mr-2" />
                Contact Sales
              </button>
            </div>
          </motion.div>
        </div>
      </section>

      {/* Footer */}
      <footer className="py-12 px-6 border-t border-gray-800">
        <div className="max-w-7xl mx-auto">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-8 mb-8">
            <div>
              <div className="flex items-center space-x-2 mb-4">
                <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center font-bold">
                  W
                </div>
                <span className="text-xl font-bold">WUI</span>
              </div>
              <p className="text-gray-400">
                Next-generation proxy management panel.
              </p>
            </div>

            <div>
              <h4 className="font-semibold mb-4">Product</h4>
              <ul className="space-y-2 text-gray-400">
                <li><a href="#features" className="hover:text-white transition">Features</a></li>
                <li><a href="#pricing" className="hover:text-white transition">Pricing</a></li>
                <li><a href="#" className="hover:text-white transition">Documentation</a></li>
                <li><a href="#" className="hover:text-white transition">API</a></li>
              </ul>
            </div>

            <div>
              <h4 className="font-semibold mb-4">Company</h4>
              <ul className="space-y-2 text-gray-400">
                <li><a href="#" className="hover:text-white transition">About</a></li>
                <li><a href="#" className="hover:text-white transition">Blog</a></li>
                <li><a href="#" className="hover:text-white transition">Careers</a></li>
                <li><a href="#" className="hover:text-white transition">Contact</a></li>
              </ul>
            </div>

            <div>
              <h4 className="font-semibold mb-4">Connect</h4>
              <div className="flex space-x-4">
                <a href="#" className="text-gray-400 hover:text-white transition">
                  <Github className="w-6 h-6" />
                </a>
                <a href="#" className="text-gray-400 hover:text-white transition">
                  <Twitter className="w-6 h-6" />
                </a>
              </div>
            </div>
          </div>

          <div className="pt-8 border-t border-gray-800 text-center text-gray-400">
            <p>&copy; {new Date().getFullYear()} WUI. All rights reserved.</p>
          </div>
        </div>
      </footer>
    </div>
  )
}
