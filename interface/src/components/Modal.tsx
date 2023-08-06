import {
  Modal,
  ModalOverlay,
  ModalContent,
  ModalHeader,
  ModalFooter,
  ModalBody,
  ModalCloseButton,
  useDisclosure,
  Text,
} from '@chakra-ui/react'
import React from 'react'

type ModalProps = {
  children: React.ReactNode
  title: string
  footer?: string
  isOpen: boolean
  onClose: () => void
}
function ModalWrapper({
  children,
  isOpen,
  onClose,
  footer,
  title,
}: ModalProps) {
  const finalRef = React.useRef(null)

  return (
    <>
      <Modal finalFocusRef={finalRef} isOpen={isOpen} onClose={onClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalHeader>{title}</ModalHeader>
          <ModalCloseButton />
          <ModalBody>{children}</ModalBody>
          <ModalFooter>
            <Text> {footer}</Text>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </>
  )
}

export default ModalWrapper
