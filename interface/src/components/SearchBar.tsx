import { useState, useRef } from "react"
import {
  Box,
  Input,
  useDisclosure,
  Modal,
  ModalOverlay,
  ModalContent,
  ModalBody,
  InputGroup,
  InputLeftElement,
} from "@chakra-ui/react"
import { useEventListener } from "@chakra-ui/hooks"
import { IoSearch } from "react-icons/io5"

function SearchBar() {
  const { isOpen, onOpen, onClose } = useDisclosure()
  const [query, setQuery] = useState("")
  const inputRef = useRef<HTMLInputElement>(null)

  const handleKeyPress = (event: KeyboardEvent) => {
    if (event.key === "Escape") {
      onClose()
    }
    if (event.key === "Enter") {
      handleSubmit()
    }
  }

  const handleSubmit = () => {
    // Perform search logic here
    console.log("Search query:", query)
  }

  useEventListener("keydown", handleKeyPress)

  return (
    <Box textAlign="center">
      <InputGroup>
        <InputLeftElement
          pointerEvents="none"
          children={<IoSearch color="gray.500" />}
        />
        <Input
          ref={inputRef}
          placeholder="Search tracks"
          onClick={() => {
            onOpen()
            inputRef.current?.blur()
          }}
          onChange={() => ({})}
          value={""}
        />
      </InputGroup>

      <Modal isOpen={isOpen} onClose={onClose}>
        <ModalOverlay />
        <ModalContent>
          <ModalBody>
            <InputGroup>
              <InputLeftElement
                pointerEvents="none"
                children={<IoSearch color="gray.500" />}
              />
              <Input
                placeholder="Search tracks"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                focusBorderColor="none"
              />
            </InputGroup>
          </ModalBody>
        </ModalContent>
      </Modal>
    </Box>
  )
}

export default SearchBar
