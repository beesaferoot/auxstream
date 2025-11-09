import { Box } from '@chakra-ui/react'
import Header from './Header'

interface LayoutProps {
  children: React.ReactNode
}

const Layout = ({ children }: LayoutProps) => {
  return (
    <Box minH="100vh" bg="gray.50">
      <Header />
      <Box as="main">{children}</Box>
    </Box>
  )
}

export default Layout

