import { useState } from 'react'
import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalCloseButton,
  Button,
  VStack,
  Input,
  FormControl,
  FormLabel,
  FormErrorMessage,
  Text,
  useToast,
  Tabs,
  TabList,
  TabPanels,
  Tab,
  TabPanel,
} from '@chakra-ui/react'
import { login as apiLogin, register as apiRegister } from '../utils/api'
import { useAuth } from '../context/AuthContext'

interface AuthModalProps {
  isOpen: boolean
  onClose: () => void
  defaultTab?: number
}

const AuthModal = ({ isOpen, onClose, defaultTab = 0 }: AuthModalProps) => {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [errors, setErrors] = useState<{ email?: string; password?: string }>({})
  const toast = useToast()
  const { login: authLogin, checkAuth } = useAuth()

  const validateEmail = (email: string): boolean => {
    const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    return re.test(email)
  }

  const validateForm = (): boolean => {
    const newErrors: { email?: string; password?: string } = {}

    if (!email.trim()) {
      newErrors.email = 'Email is required'
    } else if (!validateEmail(email)) {
      newErrors.email = 'Invalid email format'
    }

    if (!password) {
      newErrors.password = 'Password is required'
    } else if (password.length < 6) {
      newErrors.password = 'Password must be at least 6 characters'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleLogin = async () => {
    if (!validateForm()) return

    setIsLoading(true)
    try {
      const response = await apiLogin({ email, password })
      authLogin(response.data.access_token, email)
      checkAuth()

      toast({
        title: 'Login successful!',
        description: `Welcome back, ${email}`,
        status: 'success',
        duration: 3000,
        isClosable: true,
      })

      // Reset form
      setEmail('')
      setPassword('')
      setErrors({})
      onClose()
    } catch (error) {
      toast({
        title: 'Login failed',
        description: error instanceof Error ? error.message : 'Invalid credentials',
        status: 'error',
        duration: 5000,
        isClosable: true,
      })
    } finally {
      setIsLoading(false)
    }
  }

  const handleRegister = async () => {
    if (!validateForm()) return

    setIsLoading(true)
    try {
      await apiRegister({ email, password })

      toast({
        title: 'Registration successful!',
        description: 'You can now log in with your credentials',
        status: 'success',
        duration: 5000,
        isClosable: true,
      })

      // Auto-login after registration
      await handleLogin()
    } catch (error) {
      toast({
        title: 'Registration failed',
        description:
          error instanceof Error ? error.message : 'Unable to create account',
        status: 'error',
        duration: 5000,
        isClosable: true,
      })
    } finally {
      setIsLoading(false)
    }
  }

  const handleClose = () => {
    setEmail('')
    setPassword('')
    setErrors({})
    onClose()
  }

  return (
    <Modal isOpen={isOpen} onClose={handleClose} size="md">
      <ModalOverlay />
      <ModalContent>
        <ModalHeader>Welcome to AuxStream</ModalHeader>
        <ModalCloseButton />
        <ModalBody pb={6}>
          <Tabs defaultIndex={defaultTab} colorScheme="brand">
            <TabList>
              <Tab>Login</Tab>
              <Tab>Register</Tab>
            </TabList>

            <TabPanels>
              {/* Login Tab */}
              <TabPanel>
                <VStack spacing={4} align="stretch">
                  <FormControl isInvalid={!!errors.email}>
                    <FormLabel>Email</FormLabel>
                    <Input
                      type="email"
                      placeholder="Enter your email"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      isDisabled={isLoading}
                    />
                    {errors.email && (
                      <FormErrorMessage>{errors.email}</FormErrorMessage>
                    )}
                  </FormControl>

                  <FormControl isInvalid={!!errors.password}>
                    <FormLabel>Password</FormLabel>
                    <Input
                      type="password"
                      placeholder="Enter your password"
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      onKeyPress={(e) => e.key === 'Enter' && handleLogin()}
                      isDisabled={isLoading}
                    />
                    {errors.password && (
                      <FormErrorMessage>{errors.password}</FormErrorMessage>
                    )}
                  </FormControl>

                  <Button
                    colorScheme="brand"
                    onClick={handleLogin}
                    isLoading={isLoading}
                    loadingText="Logging in..."
                    size="lg"
                    mt={2}
                  >
                    Login
                  </Button>

                  <Text fontSize="sm" color="gray.600" textAlign="center">
                    Don't have an account? Switch to Register tab.
                  </Text>
                </VStack>
              </TabPanel>

              {/* Register Tab */}
              <TabPanel>
                <VStack spacing={4} align="stretch">
                  <FormControl isInvalid={!!errors.email}>
                    <FormLabel>Email</FormLabel>
                    <Input
                      type="email"
                      placeholder="Enter your email"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      isDisabled={isLoading}
                    />
                    {errors.email && (
                      <FormErrorMessage>{errors.email}</FormErrorMessage>
                    )}
                  </FormControl>

                  <FormControl isInvalid={!!errors.password}>
                    <FormLabel>Password</FormLabel>
                    <Input
                      type="password"
                      placeholder="Create a password (min. 6 characters)"
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      onKeyPress={(e) => e.key === 'Enter' && handleRegister()}
                      isDisabled={isLoading}
                    />
                    {errors.password && (
                      <FormErrorMessage>{errors.password}</FormErrorMessage>
                    )}
                  </FormControl>

                  <Button
                    colorScheme="brand"
                    onClick={handleRegister}
                    isLoading={isLoading}
                    loadingText="Creating account..."
                    size="lg"
                    mt={2}
                  >
                    Create Account
                  </Button>

                  <Text fontSize="sm" color="gray.600" textAlign="center">
                    Already have an account? Switch to Login tab.
                  </Text>
                </VStack>
              </TabPanel>
            </TabPanels>
          </Tabs>
        </ModalBody>
      </ModalContent>
    </Modal>
  )
}

export default AuthModal

