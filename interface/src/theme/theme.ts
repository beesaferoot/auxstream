import { extendTheme, type ThemeConfig } from '@chakra-ui/react'

const config: ThemeConfig = {
  initialColorMode: 'light',
  useSystemColorMode: false,
}

const theme = extendTheme({
  config,
  colors: {
    brand: {
      50: '#e6f2ff',
      100: '#bfdeff',
      200: '#99caff',
      300: '#73b6ff',
      400: '#4da2ff',
      500: '#268eff',
      600: '#0072e6',
      700: '#0059b3',
      800: '#004080',
      900: '#00264d',
    },
    accent: {
      50: '#f5e6ff',
      100: '#e0b3ff',
      200: '#cc80ff',
      300: '#b84dff',
      400: '#a31aff',
      500: '#8a00e6',
      600: '#6b00b3',
      700: '#4d0080',
      800: '#2e004d',
      900: '#10001a',
    },
  },
  fonts: {
    heading: `'Inter', -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif`,
    body: `'Inter', -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif`,
  },
  styles: {
    global: {
      body: {
        bg: 'gray.50',
        color: 'gray.800',
      },
    },
  },
  components: {
    Button: {
      baseStyle: {
        fontWeight: 'medium',
        borderRadius: 'lg',
      },
      variants: {
        solid: {
          bg: 'brand.500',
          color: 'white',
          _hover: {
            bg: 'brand.600',
            transform: 'translateY(-2px)',
            shadow: 'lg',
          },
          _active: {
            bg: 'brand.700',
          },
          transition: 'all 0.2s',
        },
        ghost: {
          _hover: {
            bg: 'gray.100',
          },
        },
      },
    },
    Card: {
      baseStyle: {
        container: {
          borderRadius: 'xl',
          overflow: 'hidden',
          transition: 'all 0.3s',
          _hover: {
            transform: 'translateY(-4px)',
            shadow: 'xl',
          },
        },
      },
    },
    Input: {
      variants: {
        outline: {
          field: {
            borderRadius: 'lg',
            _focus: {
              borderColor: 'brand.500',
              boxShadow: '0 0 0 1px var(--chakra-colors-brand-500)',
            },
          },
        },
      },
    },
    Modal: {
      baseStyle: {
        dialog: {
          borderRadius: 'xl',
        },
      },
    },
  },
})

export default theme
