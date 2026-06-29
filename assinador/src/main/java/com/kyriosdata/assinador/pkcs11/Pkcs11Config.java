package com.kyriosdata.assinador.pkcs11;

/**
 * Configuração de acesso a um dispositivo criptográfico (token USB / smart card)
 * via interface PKCS#11 (US-02.5).
 *
 * <p>Reúne os parâmetros necessários para configurar o provider
 * {@code SunPKCS11} embutido no JDK e monta a string de configuração inline
 * aceita por {@link java.security.Provider#configure(String)} a partir do
 * Java 9 (string iniciada por {@code "--"}). O PIN do dispositivo NÃO é
 * guardado aqui: é fornecido separadamente no momento da autenticação
 * (campo {@code token} da requisição), evitando retê-lo em memória além do
 * necessário.</p>
 */
public final class Pkcs11Config {

    /** Nome padrão do provider quando não informado (vira {@code SunPKCS11-Runner}). */
    public static final String DEFAULT_NAME = "Runner";

    private final String name;
    private final String library;
    private final Integer slot;

    /**
     * @param name    sufixo do nome do provider; {@code null}/vazio usa {@link #DEFAULT_NAME}
     * @param library caminho da biblioteca nativa PKCS#11 (ex.: {@code libsofthsm2.so},
     *                {@code eToken.dll}); obrigatório
     * @param slot    id do slot do dispositivo; {@code null} usa o primeiro slot
     *                disponível ({@code slotListIndex = 0})
     * @throws IllegalArgumentException se {@code library} for nulo ou em branco
     */
    public Pkcs11Config(String name, String library, Integer slot) {
        if (library == null || library.isBlank()) {
            throw new IllegalArgumentException(
                "library: caminho da biblioteca PKCS#11 obrigatório (ex.: --pkcs11-lib /usr/lib/softhsm/libsofthsm2.so)");
        }
        this.name = (name == null || name.isBlank()) ? DEFAULT_NAME : name.trim();
        this.library = library.trim();
        this.slot = slot;
    }

    public String name() {
        return name;
    }

    public String library() {
        return library;
    }

    /** Id do slot, ou {@code null} quando se usa o primeiro slot disponível. */
    public Integer slot() {
        return slot;
    }

    /**
     * Monta a configuração inline do provider {@code SunPKCS11}.
     *
     * <p>Formato aceito por {@code Provider.configure} (Java 9+): a string
     * começa com {@code "--"} e cada diretiva fica em uma linha. Quando o
     * slot não é informado, usa {@code slotListIndex = 0} (primeiro slot).</p>
     */
    public String toProviderConfig() {
        StringBuilder sb = new StringBuilder();
        sb.append("--name = ").append(name).append('\n');
        sb.append("library = ").append(library).append('\n');
        if (slot != null) {
            sb.append("slot = ").append(slot).append('\n');
        } else {
            sb.append("slotListIndex = 0").append('\n');
        }
        return sb.toString();
    }
}
