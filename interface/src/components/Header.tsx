import {
  Box,
  Flex,
  HStack,
  IconButton,
  Button,
  useDisclosure,
  Stack,
  Container,
  Text,
  Menu,
  MenuButton,
  MenuList,
  MenuItem,
  Avatar,
  MenuDivider,
} from '@chakra-ui/react'
import {
  IoMusicalNotes,
  IoMenu,
  IoClose,
  IoCloudUpload,
  IoLogIn,
  IoPersonCircle,
  IoSettings,
  IoLogOut,
} from 'react-icons/io5'
import { Link as RouterLink, useLocation, useNavigate } from 'react-router-dom'
import UploadTrackModal from './UploadTrackModal'
import BulkUploadModal from './BulkUploadModal'
import AuthModal from './AuthModal'
import { useAuth } from '../context/AuthContext'

const Header = () => {
  const { isOpen, onOpen, onClose } = useDisclosure()
  const { isOpen: isUploadOpen, onOpen: onUploadOpen, onClose: onUploadClose } = useDisclosure()
  const { isOpen: isBulkUploadOpen, onOpen: onBulkUploadOpen, onClose: onBulkUploadClose } = useDisclosure()
  const { isOpen: isAuthOpen, onOpen: onAuthOpen, onClose: onAuthClose } = useDisclosure()
  const location = useLocation()
  const navigate = useNavigate()
  const { isAuthenticated, userEmail, logout } = useAuth()

  const NavLink = ({ to, children }: { to: string; children: React.ReactNode }) => {
    const isActive = location.pathname === to
    return (
      <Button
        as={RouterLink}
        to={to}
        variant="ghost"
        colorScheme={isActive ? 'blue' : 'gray'}
        fontWeight={isActive ? 'bold' : 'medium'}
        _hover={{
          bg: 'gray.100',
          transform: 'translateY(-2px)',
          shadow: 'sm',
        }}
        transition="all 0.2s"
      >
        {children}
      </Button>
    )
  }

  return (
    <Box
      bg="white"
      px={4}
      boxShadow="sm"
      position="sticky"
      top={0}
      zIndex={1000}
      borderBottom="1px"
      borderColor="gray.200"
    >
      <Container maxW="container.xl">
        <Flex h={16} alignItems="center" justifyContent="space-between">
          {/* Logo */}
          <HStack spacing={3} as={RouterLink} to="/" _hover={{ opacity: 0.8 }}>
            <Box
              bg="gradient-to-r from-blue-500 to-purple-600"
              p={2}
              borderRadius="lg"
              display="flex"
              alignItems="center"
            >
              <IoMusicalNotes size={24} color="white" />
            </Box>
            <Text
              fontSize="xl"
              fontWeight="bold"
              bgGradient="linear(to-r, blue.500, purple.600)"
              bgClip="text"
            >
              AuxStream
            </Text>
          </HStack>

          {/* Desktop Navigation */}
          <HStack spacing={4} display={{ base: 'none', md: 'flex' }}>
            <NavLink to="/">Home</NavLink>
            <NavLink to="/player">Player</NavLink>
            <NavLink to="/trending">Trending</NavLink>
            {isAuthenticated && (
              <Menu>
                <MenuButton
                  as={Button}
                  leftIcon={<IoCloudUpload />}
                  colorScheme="brand"
                  size="sm"
                  _hover={{
                    transform: 'translateY(-2px)',
                    shadow: 'md',
                  }}
                  transition="all 0.2s"
                >
                  Upload
                </MenuButton>
                <MenuList>
                  <MenuItem
                    icon={<IoCloudUpload size={18} />}
                    onClick={onUploadOpen}
                  >
                    Single Track
                  </MenuItem>
                  <MenuItem
                    icon={<IoCloudUpload size={18} />}
                    onClick={onBulkUploadOpen}
                  >
                    Multiple Tracks
                  </MenuItem>
                </MenuList>
              </Menu>
            )}
          </HStack>

          {/* Right Side Actions */}
          <HStack spacing={4}>
            {isAuthenticated ? (
              <Menu>
                <MenuButton
                  as={IconButton}
                  rounded="full"
                  variant="ghost"
                  cursor="pointer"
                  minW={0}
                  display={{ base: 'none', md: 'flex' }}
                >
                  <Avatar
                    size="sm"
                    name={userEmail || 'User'}
                    bg="brand.500"
                  />
                </MenuButton>
                <MenuList>
                  <MenuItem
                    icon={<IoPersonCircle size={20} />}
                    onClick={() => navigate('/profile')}
                  >
                    Profile
                  </MenuItem>
                  <MenuItem
                    icon={<IoSettings size={20} />}
                    onClick={() => navigate('/settings')}
                  >
                    Settings
                  </MenuItem>
                  <MenuDivider />
                  <MenuItem
                    icon={<IoLogOut size={20} />}
                    onClick={logout}
                    color="red.500"
                  >
                    Sign Out
                  </MenuItem>
                </MenuList>
              </Menu>
            ) : (
              <Button
                leftIcon={<IoLogIn />}
                colorScheme="brand"
                size="sm"
                onClick={onAuthOpen}
                display={{ base: 'none', md: 'flex' }}
              >
                Login
              </Button>
            )}

            {/* Mobile menu button */}
            <IconButton
              size="md"
              icon={isOpen ? <IoClose /> : <IoMenu />}
              aria-label="Open Menu"
              display={{ md: 'none' }}
              onClick={isOpen ? onClose : onOpen}
            />
          </HStack>
        </Flex>

        {/* Mobile Navigation */}
        {isOpen && (
          <Box pb={4} display={{ md: 'none' }}>
            <Stack as="nav" spacing={2}>
              <NavLink to="/">Home</NavLink>
              <NavLink to="/player">Player</NavLink>
              <NavLink to="/trending">Trending</NavLink>
              {isAuthenticated ? (
                <>
                  <Button
                    leftIcon={<IoCloudUpload />}
                    colorScheme="brand"
                    justifyContent="flex-start"
                    onClick={onUploadOpen}
                  >
                    Upload Single Track
                  </Button>
                  <Button
                    leftIcon={<IoCloudUpload />}
                    colorScheme="brand"
                    variant="outline"
                    justifyContent="flex-start"
                    onClick={onBulkUploadOpen}
                  >
                    Upload Multiple Tracks
                  </Button>
                  <Button
                    variant="ghost"
                    justifyContent="flex-start"
                    leftIcon={<IoPersonCircle />}
                    onClick={() => {
                      navigate('/profile')
                      onClose()
                    }}
                  >
                    Profile
                  </Button>
                  <Button
                    variant="ghost"
                    justifyContent="flex-start"
                    leftIcon={<IoSettings />}
                    onClick={() => {
                      navigate('/settings')
                      onClose()
                    }}
                  >
                    Settings
                  </Button>
                  <Button
                    variant="ghost"
                    justifyContent="flex-start"
                    leftIcon={<IoLogOut />}
                    onClick={() => {
                      logout()
                      onClose()
                    }}
                    color="red.500"
                  >
                    Sign Out
                  </Button>
                </>
              ) : (
                <Button
                  leftIcon={<IoLogIn />}
                  colorScheme="brand"
                  justifyContent="flex-start"
                  onClick={() => {
                    onAuthOpen()
                    onClose()
                  }}
                >
                  Login
                </Button>
              )}
            </Stack>
          </Box>
        )}
      </Container>

      {/* Upload Modals */}
      <UploadTrackModal isOpen={isUploadOpen} onClose={onUploadClose} />
      <BulkUploadModal isOpen={isBulkUploadOpen} onClose={onBulkUploadClose} />

      {/* Auth Modal */}
      <AuthModal isOpen={isAuthOpen} onClose={onAuthClose} />
    </Box>
  )
}

export default Header

